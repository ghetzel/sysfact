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
	`${domain}`,
	`${hostname}`,
	`${hostname}.${domain}`,
	`${fqdn}`,
	`${uuid}`,
}

type RenderOptions struct {
	SourceDir          string
	DestDir            string `default:"~"`
	DefaultDirMode     int    `default:"493"`
	DefaultFileMode    int    `default:"420"`
	Owner              string
	Group              string
	DryRun             bool
	FollowSymlinks     bool
	AdditionalPatterns []string
	report             map[string]interface{}
}

func (self *RenderOptions) destPath(srcpath string) string {
	var relsrc = strings.TrimPrefix(srcpath, self.SourceDir)
	var dstpath = filepath.Join(self.DestDir, relsrc)

	if len(self.report) > 0 {
		dstpath = strings.ReplaceAll(dstpath, `[[`, `${`)
		dstpath = strings.ReplaceAll(dstpath, `]]`, `}`)

		return maputil.M(self.report).Sprintf(dstpath)
	} else {
		return dstpath
	}
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

		var visited = make(map[string]bool)
		var report = maputil.M(r)

		options.report = r

		for _, srcDirPattern := range append(RenderPatterns, options.AdditionalPatterns...) {
			var srcdir = report.Sprintf(srcDirPattern)
			srcdir = filepath.Join(basedir, srcdir)
			srcdir = fileutil.MustExpandUser(srcdir)

			// make sure the directory exists
			if fileutil.DirExists(srcdir) {
				// and that we haven't already been here
				if _, seen := visited[srcdir]; !seen {
					visited[srcdir] = true
					options.SourceDir = srcdir

					if err := renderTree(options); err != nil {
						return err
					}
				}
			}
		}

		return nil
	} else {
		return err
	}
}

// Recursively copies all files in RenderOptions.SourceDir into RenderOptions.DestDir, creating any
// intermediate directories as necessary.  Identifies and renders templates into text files in DestDir,
// as well as templated filenames.
func renderTree(options *RenderOptions) error {
	options.DestDir = fileutil.MustExpandUser(options.DestDir)

	// ensure existence of source and destination directories, and enforces permissions on the destination.
	if !fileutil.DirExists(options.SourceDir) {
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

	// recursively walk all files in SourceDir, copying, rendering, and following as necessary.
	if err := filepath.Walk(options.SourceDir, func(srcpath string, info os.FileInfo, err error) error {
		var dstpath = options.destPath(srcpath)
		var mode = options.ModeFor(info)
		var source io.ReadCloser
		var verb string = `create`
		var linkTarget string

		// if we're not following symlinks, we'll need the actual target said symlink.  linkTarget
		// will remain empty if this file is not a symlink.
		if !options.FollowSymlinks {
			if t, err := os.Readlink(srcpath); err == nil {
				linkTarget = t
				verb = `  link`
			}
		}

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

			// we got a template.  Render it using Diecast and set the source to a buffer containing the rendered output
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

		// create a symlink with the same target as the source one we've got
		if linkTarget != `` {
			if fileutil.Exists(dstpath) {
				if err := os.Remove(dstpath); err != nil {
					return err
				}
			}

			if target, err := os.Readlink(srcpath); err == nil {
				if err := os.Symlink(target, dstpath); err == nil {
					options.log("%s | %v | %v -> %v", verb, mode, dstpath, target)
					return nil
				} else {
					return err
				}
			} else {
				return err
			}
		} else if source != nil {
			defer source.Close()

			if options.DryRun {
				options.log("%s | %v | %v", verb, mode, dstpath)
			} else {
				if fileutil.Exists(dstpath) {
					if err := os.Remove(dstpath); err != nil {
						return err
					}
				}

				// create destination file
				if dest, err := os.Create(dstpath); err == nil {
					defer dest.Close()

					// copy source buffer
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

				// enforce permissions
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
