// Package fileutil file util.
package fileutil

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// WriteStringBuilderToFile write string
func WriteStringBuilderToFile(builder *strings.Builder, tarPath, name string) error {
	fullPath := path.Join(tarPath, name)

	err := os.MkdirAll(tarPath, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(builder.String())
	if err != nil {
		return err
	}

	return nil
}

// CreatePath create path.
func CreatePath(pathStr string) error {
	isExist, err := PathExists(pathStr)
	if err != nil {
		return err
	}
	if isExist {
		return nil
	}

	err = os.MkdirAll(pathStr, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// PathExists path exists.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// RemoveContents remove contents.
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		if strings.Contains(err.Error(), " no such file or directory") {
			return nil
		}
		return err
	}
	defer func() {
		_ = d.Close()
	}()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveDir remove dir.
func RemoveDir(dir string) error {
	isExist, err := PathExists(dir)
	if err != nil {
		return err
	}
	if !isExist {
		return nil
	}

	err = RemoveContents(dir)
	if err != nil {
		return err
	}

	err = os.Remove(dir)
	if err != nil {
		return err
	}

	return nil
}

// ZipDir zip dir
func ZipDir(dir, zipFile string) error {

	fz, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = fz.Close()
	}()

	zw := zip.NewWriter(fz)
	defer func() {
		_ = zw.Close()
	}()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("file is empty path is %s ", path)
		}

		if !info.IsDir() {
			fDest, err := zw.Create(path[len(dir)+1:])
			if err != nil {
				return nil
			}
			fSrc, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				_ = fSrc.Close()
			}()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// UnzipDir unzip dir.
func UnzipDir(zipFile, dir string) {

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Fatalf("Open zip file failed: %s \n", err.Error())
	}
	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		func() {
			path := dir + string(filepath.Separator) + f.Name
			_ = os.MkdirAll(filepath.Dir(path), 0o755)
			fDest, err := os.Create(path)
			if err != nil {
				log.Printf("Create failed: %s\n", err.Error())
				return
			}
			defer func() {
				_ = fDest.Close()
			}()

			fSrc, err := f.Open()
			if err != nil {
				log.Printf("Open failed: %s\n", err.Error())
				return
			}
			defer func() {
				_ = fSrc.Close()
			}()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Printf("Copy failed: %s\n", err.Error())
				return
			}
		}()
	}
}

// LineCounter line counter.
func LineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// FileLineCounter file line.
func FileLineCounter(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	return LineCounter(file)
}

// DownloadFile download.
func DownloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("redis tool get err", err.Error())
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	out, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("redis tool copy err", err.Error())
		return err
	}

	return nil
}

// EnsurePath unsure path.
func EnsurePath(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			return
		}
		log.Println(err)
		panic("failed to create path:" + path)
	}
}

// UnGz ungz.
func UnGz(srcGz, dstFile string) error {
	fr, err := os.Open(srcGz)
	if err != nil {
		return err
	}
	defer func() {
		_ = fr.Close()
	}()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer func() {
		_ = gr.Close()
	}()

	isExist, err := PathExists(dstFile)
	if err != nil {
		return err
	}

	var fw *os.File
	if isExist {
		fw, err = os.OpenFile(dstFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o666)
	} else {
		fw, err = os.Create(dstFile)
	}
	if err != nil {
		return err
	}
	defer func() {
		_ = fw.Close()
	}()

	// 写文件
	_, err = io.Copy(fw, gr)
	if err != nil {
		return err
	}

	// 删除gz压缩文件
	err = os.Remove(srcGz)
	if err != nil {
		return err
	}
	return nil
}

// Compress1 compress1
// TODO 不完善  请勿使用
func Compress1(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer func() {
		_ = d.Close()
	}()
	gw := gzip.NewWriter(d)
	defer func() {
		_ = gw.Close()
	}()

	for _, file := range files {
		err := Compress(file, "", gw)
		if err != nil {
			return err
		}
	}
	return nil
}

// Compress compress
func Compress(file *os.File, prefix string, tw *gzip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	switch {
	case info.IsDir():
		prefix = prefix + "/" + info.Name()
		fileInfos, errTmp := file.Readdir(-1)
		if errTmp != nil {
			return errTmp
		}
		for _, fi := range fileInfos {
			f, errTmp := os.Open(file.Name() + "/" + fi.Name())
			if errTmp != nil {
				return errTmp
			}
			errTmp = Compress(f, prefix, tw)
			if errTmp != nil {
				return errTmp
			}
		}
	default:
		_, err = io.Copy(tw, file)
		defer func() {
			_ = file.Close()
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// JSONOut json out.
func JSONOut(filename string, i interface{}) error {
	out, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	if out != nil {
		defer func() {
			_ = out.Close()
		}()
	}

	err = json.NewEncoder(out).Encode(i)
	if err != nil {
		return err
	}
	return nil
}

// CopyFolder 复制文件夹
func CopyFolder(source, destination string) error {
	// create dest dir
	err := os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		return err
	}

	// create source dir
	files, err := filepath.Glob(filepath.Join(source, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		// source file stat
		fileInfo, err := os.Stat(file)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			tempPath := filepath.Join(destination, fileInfo.Name())
			err1 := os.MkdirAll(destination, os.ModePerm)
			if err1 != nil {
				return err1
			}
			err1 = CopyFolder(file, tempPath)
			if err1 != nil {
				return err1
			}
		} else {
			err1 := CopyFile(file, filepath.Join(destination, fileInfo.Name()))
			if err1 != nil {
				return err1
			}
		}
	}

	return nil
}

// CopyFile copy file
func CopyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 复制文件内容
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	sourceFileInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.Chmod(destination, sourceFileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}
