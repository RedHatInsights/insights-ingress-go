package local

import (
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/redhatinsights/insights-ingress-go/stage"
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

func (l *LocalStager) Stage(in *stage.Input) (string, error) {
	f, err := ioutil.TempFile(l.WorkingDir, "stage")
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	buf, err := ioutil.ReadAll(in.Reader)
	if err != nil {
		return "", err
	}

	_, err = f.Write(buf)
	if err != nil {
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
