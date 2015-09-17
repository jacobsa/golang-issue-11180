package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/jacobsa/fuse"
	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
)

var fDelay = flag.Duration(
	"delay", time.Second, "How long to delay file creation.")

func mount(mountPoint string) (mfs *fuse.MountedFileSystem, err error) {
	// Create the file system.
	fs, err := newFileSystem(*fDelay)
	if err != nil {
		err = fmt.Errorf("newFileSystem: %v", err)
		return
	}

	// Mount the file system.
	mfs, err = fuse.Mount(
		mountPoint,
		fuseutil.NewFileSystemServer(fs),
		&fuse.MountConfig{})

	if err != nil {
		err = fmt.Errorf("Mount: %v", err)
		return
	}

	return

	// Wait for unmount.
	err = mfs.Join(context.Background())
	if err != nil {
		err = fmt.Errorf("Join: %v", err)
		return
	}

	return
}

type fileSystem struct {
	fuseutil.NotImplementedFileSystem

	delay time.Duration
	uid   uint32
	gid   uint32

	mu          sync.Mutex
	nextInodeID fuseops.InodeID
}

func newFileSystem(delay time.Duration) (fs *fileSystem, err error) {
	uid, gid, err := myUserAndGroup()
	if err != nil {
		err = fmt.Errorf("myUserAndGroup: %v", err)
		return
	}

	fs = &fileSystem{
		delay:       delay,
		uid:         uid,
		gid:         gid,
		nextInodeID: fuseops.RootInodeID + 1,
	}
	return
}

func (fs *fileSystem) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) (err error) {
	return
}

func (fs *fileSystem) CreateFile(
	ctx context.Context,
	op *fuseops.CreateFileOp) (err error) {
	time.Sleep(fs.delay)

	// Allocate an inode ID.
	fs.mu.Lock()
	id := fs.nextInodeID
	fs.nextInodeID++
	fs.mu.Unlock()

	// Always succeed.
	op.Entry = fuseops.ChildInodeEntry{
		Child: id,
		Attributes: fuseops.InodeAttributes{
			Mode: 0666,
			Uid:  fs.uid,
			Gid:  fs.gid,
		},
	}

	return
}

func (fs *fileSystem) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) (err error) {
	// The root directory looks like a directory; everything else looks like a file.
	var mode os.FileMode
	switch op.Inode {
	case fuseops.RootInodeID:
		mode = 0777 | os.ModeDir

	default:
		mode = 0666
	}

	op.Attributes = fuseops.InodeAttributes{
		Mode: mode,
		Uid:  fs.uid,
		Gid:  fs.gid,
	}

	return
}

func (fs *fileSystem) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) (err error) {
	// No name exists.
	err = fuse.ENOENT
	return
}
