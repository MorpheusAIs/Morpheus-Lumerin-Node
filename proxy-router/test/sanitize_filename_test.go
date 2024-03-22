package test

import (
	"os"
	"path"
	"testing"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
)

func TestSanitizeFilenameFilesystem(t *testing.T) {
	filename := "!@#$%^&*()-+=~`';:<>,./?\\-[::1]:8080.log"
	sanitizedFilename := lib.SanitizeFilename(filename)
	pth := path.Join(os.TempDir(), sanitizedFilename)

	file, err := os.OpenFile(pth, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Remove(pth)
	if err != nil {
		t.Fatal(err)
	}
}
