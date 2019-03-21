package tinycloud

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

var (
	errVirtualResource = errors.New("cannot resolve virtual resource")
	errNoPermission    = os.ErrPermission
)

const (
	_ int = iota
	invalid
	index
	shares
	userFolder
	shareFolder
)

// "/<user>/..."
// "/shares/<share>/..."

type resource struct {
	name string
	kind int
	root bool
}

func parseResource(ctx context.Context, name string) resource {
	var (
		user = ctx.Value(ctxUser{}).(string)
		kind = invalid
		root = true
	)

	name = path.Clean(name)

	switch name {
	case "":
		kind = invalid

	case "/":
		kind = index

	case "/shares":
		kind = shares

	default:
		parts := strings.Split(name[1:], "/")

		if len(parts) > 0 {
			if parts[0] == user {
				kind = userFolder
				root = len(parts) == 1
			} else if parts[0] == "shares" {
				// TODO: if has access to share
			}
		}
	}

	return resource{
		name: name,
		kind: kind,
		root: root,
	}
}

func (r resource) canOpen() bool {
	return r.kind == userFolder || r.kind == shareFolder
}

func (r resource) canModify() bool {
	return !r.root && r.canOpen()
}

type fs struct {
	files afero.Fs
}

func (f fs) resolve(r resource) string {
	switch r.kind {
	case userFolder:
		return path.Join("users", r.name)

	case shareFolder:
		return r.name

	default:
		panic(errVirtualResource)
	}
}

func (f fs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if r := parseResource(ctx, name); r.canModify() {
		return f.files.Mkdir(f.resolve(r), perm)
	}

	return errNoPermission
}

func (f fs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	switch r := parseResource(ctx, name); r.kind {
	case index:
		return vDir{
			name: "/",
			children: []os.FileInfo{
				vFileInfo{
					name: ctx.Value(ctxUser{}).(string),
					mode: os.ModeDir,
				},
				vFileInfo{
					name: "shares",
					mode: os.ModeDir,
				},
			},
		}, nil
	case shares:
		return vDir{name: "shares"}, nil

	default:
		if r.canOpen() {
			return f.files.OpenFile(f.resolve(r), flag, perm)
		}
	}

	return nil, errNoPermission
}

func (f fs) RemoveAll(ctx context.Context, name string) error {
	if r := parseResource(ctx, name); r.canModify() {
		return f.files.RemoveAll(f.resolve(r))
	}

	return errNoPermission
}

func (f fs) Rename(ctx context.Context, oldName, newName string) error {
	var (
		rOld = parseResource(ctx, oldName)
		rNew = parseResource(ctx, newName)
	)

	if rOld.canModify() && rNew.canModify() {
		return f.files.Rename(f.resolve(rOld), f.resolve(rNew))
	}

	return errNoPermission
}

func (f fs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	switch r := parseResource(ctx, name); r.kind {
	case index:
		return vFileInfo{
			name: "/",
			mode: os.ModeDir,
		}, nil

	case shares:
		return vFileInfo{
			name: "shares",
			mode: os.ModeDir,
		}, nil

	default:
		if r.canOpen() {
			return f.files.Stat(f.resolve(r))
		}
	}

	return nil, errNoPermission
}

type vFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (v vFileInfo) Name() string {
	return v.name
}

func (v vFileInfo) Size() int64 {
	return v.size
}

func (v vFileInfo) Mode() os.FileMode {
	return v.mode
}

func (v vFileInfo) ModTime() time.Time {
	return v.modTime
}

func (v vFileInfo) IsDir() bool {
	return v.Mode().IsDir()
}

func (v vFileInfo) Sys() interface{} {
	return nil
}

type vDir struct {
	name     string
	children []os.FileInfo
}

func (v vDir) Write([]byte) (int, error) {
	return 0, os.ErrClosed
}

func (v vDir) Close() error {
	return nil
}

func (v vDir) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (v vDir) Seek(int64, int) (int64, error) {
	return 0, os.ErrClosed
}

func (v vDir) Readdir(int) ([]os.FileInfo, error) {
	return v.children, nil
}

func (v vDir) Stat() (os.FileInfo, error) {
	return vFileInfo{
		name: v.name,
		mode: os.ModeDir,
	}, nil
}
