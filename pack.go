package nuget

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	nuspec "github.com/soloworks/go-nuspec"
)

// PackNupkg produces a .nupkg file in byte format
func PackNupkg(ns *nuspec.NuSpec, basePath string, outputPath string) ([]byte, error) {

	// Assume filename from ID
	nsfilename := ns.Meta.ID + ".nuspec"

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive
	w := zip.NewWriter(buf)
	defer w.Close()

	// Create a new Contenttypes Structure
	ct := NewContentTypes()

	// Add .nuspec to Archive
	b, err := ns.ToBytes()
	if err != nil {
		return nil, err
	}
	if err := archiveFile(filepath.Base(nsfilename), w, b); err != nil {
		return nil, err
	}
	ct.Add(filepath.Ext(nsfilename))

	// Process files
	// If there are no file globs specified then
	if len(ns.Files.File) == 0 {
		// walk the basePath and zip up all found files. Everything.]
		err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Base(path) != filepath.Base(nsfilename) {
				// Open the file
				x, err := os.Open(path)
				if err != nil {
					return err
				}
				// Gather all contents
				y, err := ioutil.ReadAll(x)
				if err != nil {
					return err
				}
				// Set relative path for file in archive
				p, err := filepath.Rel(basePath, path)
				if err != nil {
					return err
				}
				// Store the file
				if err := archiveFile(p, w, y); err != nil {
					return err
				}
				// Add extension to the Rels file
				ct.Add(filepath.Ext(p))
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// For each of the specified globs, get files an put in target
		for _, f := range ns.Files.File {
			// Apply glob, cater for
			matches, err := filepath.Glob(filepath.ToSlash(filepath.Join(basePath, f.Source)))
			if err != nil {
				return nil, err
			}
			for _, m := range matches {
				info, err := os.Stat(m)
				if !info.IsDir() && filepath.Base(m) != filepath.Base(nsfilename) {
					// Open the file
					x, err := os.Open(m)
					if err != nil {
						return nil, err
					}
					// Gather all contents
					y, err := ioutil.ReadAll(x)
					if err != nil {
						return nil, err
					}
					// Set relative path for file in archive
					p, err := filepath.Rel(basePath, m)
					if err != nil {
						return nil, err
					}
					// Overide path if Target is set
					if f.Target != "" {
						p = filepath.Join(f.Target, filepath.Base(m))
					}
					// Store the file
					if err := archiveFile(p, w, y); err != nil {
						return nil, err
					}
					// Add extension to the Rels file
					ct.Add(filepath.Ext(p))
				}
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Create and add .psmdcp file to Archive
	pf := NewPsmdcpFile()
	pf.Creator = ns.Meta.Authors
	pf.Description = ns.Meta.Description
	pf.Identifier = ns.Meta.ID
	pf.Version = ns.Meta.Version
	pf.Keywords = ns.Meta.Tags
	pf.LastModifiedBy = "go-nuget"
	b, err = pf.ToBytes()
	if err != nil {
		return nil, err
	}
	pfn := "package/services/metadata/core-properties/" + randomString(32) + ".psmdcp"
	if err := archiveFile(pfn, w, b); err != nil {
		return nil, err
	}
	ct.Add(filepath.Ext(pfn))

	// Create and add .rels to Archive
	rf := NewRelFile()
	rf.Add("http://schemas.microsoft.com/packaging/2010/07/manifest", "/"+filepath.Base(nsfilename))
	rf.Add("http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", pfn)

	b, err = rf.ToBytes()
	if err != nil {
		return nil, err
	}
	if err := archiveFile(filepath.Join("_rels", ".rels"), w, b); err != nil {
		return nil, err
	}
	ct.Add(filepath.Ext(".rels"))

	// Add [Content_Types].xml to Archive
	b, err = ct.ToBytes()
	if err != nil {
		return nil, err
	}
	if err := archiveFile(`[Content_Types].xml`, w, b); err != nil {
		return nil, err
	}

	// Close the zipwriter
	w.Close()

	// Return
	return buf.Bytes(), nil
}

func archiveFile(fn string, w *zip.Writer, b []byte) error {

	// Create the file in the zip
	f, err := w.Create(filepath.ToSlash(fn))
	if err != nil {
		return err
	}

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	if err != nil {
		return err
	}
	return nil
}

const letterBytes = "abcdef0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
