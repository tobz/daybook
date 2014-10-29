package archive

import "io"
import "os"
import "fmt"
import "path/filepath"
import "archive/tar"
import "compress/gzip"

type TarGzArchive struct {
    compressed io.ReadCloser
    closed bool
}

func NewTarGzArchive(compressed io.ReadCloser) *TarGzArchive {
    return &TarGzArchive{compressed: compressed, closed: false}
}

func (a *TarGzArchive) Extract(rootDir string) error {
    rootPath, err := filepath.Abs(rootDir)
    if err != nil {
        return err
    }

    defer func() {
        a.compressed.Close()
        a.closed = true
    }()

    gzipReader, err := gzip.NewReader(a.compressed)
    if err != nil {
        return err
    }
    defer gzipReader.Close()

    tarReader := tar.NewReader(gzipReader)

    for {
        header, err := tarReader.Next()
        if err != nil {
            if err == io.EOF {
                break
            }

            return err
        }

        err = a.handleHeader(header, tarReader, rootPath)
        if err != nil {
            return err
        }
    }

    return nil
}

func (a *TarGzArchive) handleHeader(header *tar.Header, reader *tar.Reader, rootDir string) error {
    switch header.Typeflag {
    case tar.TypeReg, tar.TypeRegA:
        return a.handleFile(header, reader, rootDir)
    case tar.TypeDir:
        return a.handleDir(header, rootDir)
    }

    return fmt.Errorf("unhandled object type '%#v'", header.Typeflag)
}

func (a *TarGzArchive) handleFile(header *tar.Header, reader *tar.Reader, rootDir string) error {
    baseDirectory := filepath.Join(rootDir, filepath.Dir(header.Name))
    fileName := filepath.Join(baseDirectory, filepath.Base(header.Name))

    if !ensureDirectory(baseDirectory) {
        return fmt.Errorf("base directory '%s' doesn't exist", baseDirectory)
    }

    f, err := os.Create(fileName)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = io.Copy(f, reader)
    if err != nil {
        return err
    }

    return ensurePermissions(fileName, header)
}

func (a *TarGzArchive) handleDir(header *tar.Header, rootDir string) error {
    directoryName := filepath.Join(rootDir, filepath.Base(header.Name))

    if !ensureDirectory(directoryName) {
        err := os.Mkdir(directoryName, os.FileMode(header.Mode) & os.ModePerm)
        if err != nil {
            return err
        }
    }

    return ensurePermissions(directoryName, header)
}

func ensurePermissions(path string, header *tar.Header) error {
    err := os.Chmod(path, os.FileMode(header.Mode) & os.ModePerm)
    if err != nil {
        return err
    }

    err = os.Chown(path, header.Uid, header.Gid)
    if err != nil {
        return err
    }

    return nil
}

func ensureDirectory(dir string) bool {
    fi, err := os.Stat(dir)
    if err == os.ErrNotExist {
        return false
    }

    if err != nil {
        return false
    }

    return fi.IsDir()
}
