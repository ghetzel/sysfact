package sysfact

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ghetzel/diecast"
	"github.com/ghetzel/go-stockutil/convutil"
	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/mcuadros/go-defaults"
)

var RenderAsTemplatePrefix = `@`
var RenderPatterns = []string{
	`any/any`,
	`any/${arch}`,
	`${os.platform}/any`,
	`${os.platform}/${arch}`,
	`${os.family}/any`,
	`${os.family}/${arch}`,
	`${os.distribution}/any`,
	`${os.distribution}/${arch}`,
	`${os.distribution}-${os.version}/any`,
	`${os.distribution}-${os.version}/${arch}`,
}

type RenderOptions struct {
	DestDir         string `default:"~"`
	DefaultDirMode  int    `default:"493"`
	DefaultFileMode int    `default:"420"`
	Owner           string
	Group           string
	DryRun          bool
	report          map[string]interface{}
}

func (self *RenderOptions) log(format string, args ...interface{}) {
	var logPrefix string

	if self.DryRun {
		logPrefix = `[dry-run] `
	}

	log.Noticef(logPrefix+format, args...)
}

func (self *RenderOptions) logMinor(format string, args ...interface{}) {
	var logPrefix string

	if self.DryRun {
		logPrefix = `[dry-run] `
	}

	log.Infof(logPrefix+format, args...)
}

func (self *RenderOptions) ModeFor(info os.FileInfo) os.FileMode {
	if info == nil || info.IsDir() {
		return os.FileMode(self.DefaultDirMode)
	} else {
		return os.FileMode(self.DefaultFileMode)
	}
}

func (self *RenderOptions) Enforce(path string) error {
	if have, err := os.Stat(path); err == nil {
		var wantMode = self.ModeFor(have)

		if wantMode.Perm() != have.Mode().Perm() {
			if self.DryRun {
				self.logMinor("chmod | %v | %v", wantMode, path)
			} else {
				if err := os.Chmod(path, wantMode.Perm()); err != nil {
					return fmt.Errorf("enforce: %v", err)
				}
			}
		}

		var uid int = -1
		var gid int = -1

		if self.Owner != `` {
			if owner, err := user.Lookup(self.Owner); err == nil {
				uid = int(typeutil.Int(owner.Uid))
			} else {
				return fmt.Errorf("enforce: %v", err)
			}
		}

		if self.Group != `` {
			if group, err := user.LookupGroup(self.Group); err == nil {
				gid = int(typeutil.Int(group.Gid))
			} else {
				return fmt.Errorf("enforce: %v", err)
			}
		}

		if uid >= 0 || gid >= 0 {
			self.logMinor("chmod | %v | %v", wantMode, path)

			if self.DryRun {
				return nil
			} else {
				return os.Chown(path, uid, gid)
			}
		} else {
			return nil
		}
	} else if self.DryRun {
		return nil
	} else {
		return err
	}
}

func Render(basedir string, options *RenderOptions) error {
	if options == nil {
		options = new(RenderOptions)
	}

	defaults.SetDefaults(options)

	if r, err := Report(); err == nil {
		r, _ = maputil.DiffuseMap(r, `.`)

		var report = maputil.M(r)

		options.report = r

		for _, srcDirPattern := range RenderPatterns {
			var srcdir = report.Sprintf(srcDirPattern)
			srcdir = filepath.Join(basedir, srcdir)
			srcdir = fileutil.MustExpandUser(srcdir)

			if fileutil.DirExists(srcdir) {
				// options.log("  walk | %s", srcdir)

				if err := renderTree(srcdir, options); err != nil {
					return err
				}
			}
		}

		return nil
	} else {
		return err
	}
}

func renderTree(srcdir string, options *RenderOptions) error {
	options.DestDir = fileutil.MustExpandUser(options.DestDir)
	// options.DestDir = strings.TrimPrefix(options.DestDir, srcdir)

	if !fileutil.DirExists(srcdir) {
		return fmt.Errorf("Must specify a source directory tree to render.")
	} else if !fileutil.DirExists(options.DestDir) {
		var mode = options.ModeFor(nil)

		if !options.DryRun {
			if err := os.MkdirAll(options.DestDir, mode); err != nil {
				return err
			}
		}

		if err := options.Enforce(options.DestDir); err != nil {
			return err
		}
	}

	if err := filepath.Walk(srcdir, func(srcpath string, info os.FileInfo, err error) error {
		var relsrc = strings.TrimPrefix(srcpath, srcdir)
		var dstpath = filepath.Join(options.DestDir, relsrc)
		var mode = options.ModeFor(info)
		var source io.ReadCloser
		var verb string = `create`

		if info.IsDir() {
			if fileutil.FileExists(dstpath) {
				log.Warningf("%s: destination exists and is a file, expecting directory", dstpath)
			} else if !fileutil.DirExists(dstpath) {
				if !options.DryRun {
					return os.MkdirAll(dstpath, mode)
				}
			}
		} else if ddir, dname := filepath.Split(dstpath); strings.HasPrefix(dname, RenderAsTemplatePrefix) {
			dstpath = filepath.Join(ddir, strings.TrimPrefix(dname, RenderAsTemplatePrefix))
			verb = `render`

			// we got a template.  Rendering it using Diecast and set the source to a buffer containing the rendered output
			if src, err := fileutil.ReadAllString(srcpath); err == nil {
				if rendered, err := diecast.EvalInline(
					src,
					options.report,
					diecast.GetStandardFunctions(nil),
				); err == nil {
					source = ioutil.NopCloser(bytes.NewBufferString(rendered))
				} else {
					return fmt.Errorf("render template: %v", err)
				}
			} else {
				return err
			}
		} else if file, err := os.Open(srcpath); err == nil {
			source = file
		} else {
			return err
		}

		if source != nil {
			defer source.Close()

			if options.DryRun {
				options.log("%s | %v | %v", verb, mode, dstpath)
			} else {
				if dest, err := os.Create(dstpath); err == nil {
					defer dest.Close()

					if n, err := io.Copy(dest, source); err == nil {
						options.log("%s | %v | %v (%v)", verb, mode, dstpath, convutil.Bytes(n))
						dest.Close()
						source.Close()
					} else {
						return err
					}
				} else if log.ErrHasSuffix(err, ` text file busy`) {
					log.Warningf("skipping %q: file locked for writing", dstpath)
					return nil
				} else {
					return err
				}

				return options.Enforce(dstpath)
			}
		}

		return nil
	}); err == nil {
		return nil
	} else {
		return err
	}
}
