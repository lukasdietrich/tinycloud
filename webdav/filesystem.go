package webdav

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/webdav"

	"github.com/lukasdietrich/tinycloud/storage"
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
	kind       int
	foldername string
	filename   string
	root       bool
}

func parseResource(ctx context.Context, name string) resource {
	var (
		user       = ctx.Value(ctxUser{}).(string)
		kind       = invalid
		foldername = ""
		filename   = path.Clean(name)
		root       = true
	)

	switch filename {
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
				foldername = user
				filename = path.Join(parts[1:]...)
				root = len(parts) == 1
			} else if parts[0] == "shares" {
				// TODO: if has access to share
			}
		}
	}

	return resource{
		kind:       kind,
		foldername: foldername,
		filename:   filename,
		root:       root,
	}
}

func (r resource) canOpen() bool {
	return r.kind == userFolder || r.kind == shareFolder
}

func (r resource) canModify() bool {
	return !r.root && r.canOpen()
}

type fs struct {
	storage *storage.Storage
}

func (f fs) resolve(r resource) string {
	switch r.kind {
	case userFolder:
		return f.storage.Resolve(storage.Users, r.foldername, r.filename)

	case shareFolder:
		return f.storage.Resolve(storage.Shares, r.foldername, r.filename)

	default:
		panic(errVirtualResource)
	}
}

func (f fs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if r := parseResource(ctx, name); r.canModify() {
		return f.storage.Mkdir(f.resolve(r), perm)
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
			return f.storage.OpenFile(f.resolve(r), flag, perm)
		}
	}

	return nil, errNoPermission
}

func (f fs) RemoveAll(ctx context.Context, name string) error {
	if r := parseResource(ctx, name); r.canModify() {
		return f.storage.RemoveAll(f.resolve(r))
	}

	return errNoPermission
}

func (f fs) Rename(ctx context.Context, oldName, newName string) error {
	var (
		rOld = parseResource(ctx, oldName)
		rNew = parseResource(ctx, newName)
	)

	if rOld.canModify() && rNew.canModify() {
		return f.storage.Rename(f.resolve(rOld), f.resolve(rNew))
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
			return f.storage.Stat(f.resolve(r))
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
