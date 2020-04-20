package functions

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"test.com/video/types"
)

var containers []types.Container
var configs map[string]*types.Config

func FindContainerByExtension(search string) (string, error) {
	if containers == nil {
		path, err := os.Getwd()
		if err != nil {
			return "", err
		}

		file, err := os.Open(filepath.Join(path, "config", "containers.yaml"))
		if err != nil {
			return "", err
		}

		content, err := ioutil.ReadAll(file)
		if err != nil {
			return "", err
		}

		err = yaml.Unmarshal(content, &containers)
		if err != nil {
			return "", err
		}
	}

	res := ""
	for _, container := range containers {
		contains := false
		for _, extension := range container.Extensions {
			if extension == search {
				res = container.Name
				contains = true
				break
			}
		}

		if contains {
			break
		}
	}
	return res, nil
}

func FindConfigByContainer(containerName string) (*types.Config, error) {
	if len(configs) == 0 {
		configs = make(map[string]*types.Config)
	}

	if configs[containerName] == nil {
		path, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		file, err := os.Open(filepath.Join(path, "config", containerName+".yaml"))
		if err != nil {
			return nil, err
		}

		content, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		config := new(types.Config)
		err = yaml.Unmarshal(content, config)
		if err != nil {
			return nil, err
		}
		configs[containerName] = config
	}

	return configs[containerName], nil
}

func FindConfigByExtension(extension string) (*types.Config, error) {
	container, err := FindContainerByExtension(extension)
	if err != nil {
		return nil, err
	}

	return FindConfigByContainer(container)
}

func ErrProcess(err error, c *gin.Context) bool {
	return ErrProcessCallback(err, c, nil)
}

func ErrProcessCallback(err error, c *gin.Context, callback func(err error) bool) bool {
	if err != nil && (callback == nil || callback(err)) {
		c.String(http.StatusInternalServerError, "Server encountered an error")
		return true
	}
	return false
}
