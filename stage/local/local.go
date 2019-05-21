package local

import (
	"cloud.redhat.com/ingress/stage"
	"io/ioutil"
	"log"
	"os"
	"net/url"
)

func New(workingDir string) *LocalStager {
	d, err := ioutil.TempDir(workingDir, "ingress")
	if err != nil {
		panic("Failed to create a tmp dir: " + err.Error())
	}
	return &LocalStager{
		WorkingDir: d,
	}
}

func cleanUp(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func (l *LocalStager) Stage(in *stage.Input) (string, error) {
	f, err := ioutil.TempFile(l.WorkingDir, "stage")
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(in.Reader)
	if err != nil {
		cleanUp(f)
		return "", err
	}

	_, err = f.Write(buf)
	if err != nil {
		cleanUp(f)
		return "", nil
	}

	return "file://" + f.Name(), nil
}

func (l *LocalStager) Reject(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	os.Remove(u.Path)
	return nil
}

// CleanUp removes the temp directory used by LocalStager
func (l *LocalStager) CleanUp() {
	err := os.RemoveAll(l.WorkingDir)
	if err != nil {
		log.Printf("Attempted to clean up working directory but failed: %v", err)
	}
}
