package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func wellknown(c *gin.Context) {
	c.JSON(200, gin.H{
		"providers.v1": "/v1/providers/",
	})
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func listVersions(c *gin.Context) {
	ns, ok := c.Params.Get("namespace")
	if !ok {
		c.JSON(404, "namespace param is required")
		return
	}

	name, ok := c.Params.Get("name")
	if !ok {
		c.JSON(404, "name param is required")
		return
	}

	file_names := dirwalk(fmt.Sprintf("provider/%s/%s", ns, name))

	versions := make([]map[string]any, len(file_names))
	for i, n := range file_names {
		rep := regexp.MustCompile(`.json$`)
		e := filepath.Base(rep.ReplaceAllString(n, ""))
		versions[i] = map[string]any{
			"version": strings.Split(e, "_")[2],
		}
	}

	c.JSON(200, gin.H{
		"versions": versions,
	})
}

func download(c *gin.Context) {
	ns, ok := c.Params.Get("namespace")
	if !ok {
		c.JSON(404, "namespace param is required")
		return
	}

	name, ok := c.Params.Get("name")
	if !ok {
		c.JSON(404, "name param is required")
		return
	}

	version, ok := c.Params.Get("version")
	if !ok {
		c.JSON(404, "version param is required")
		return
	}

	os_val, ok := c.Params.Get("os")
	if !ok {
		c.JSON(404, "os param is required")
		return
	}

	archi, ok := c.Params.Get("archi")
	if !ok {
		c.JSON(404, "archi param is required")
		return
	}

	b, err := os.ReadFile(fmt.Sprintf("provider/%s/%s/%s_%s_%s.json", ns, name, os_val, archi, version))
	if err != nil {
		log.Println(err)
		c.JSON(404, err)
		return
	}

	res := map[string]any{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Println(err)
		c.JSON(500, err)
		return
	}

	c.JSON(200, res)
}

func regist(c *gin.Context) {
	id := os.Getenv("PGP_ID")
	if id == "" {
		c.JSON(500, gin.H{
			"message": "PGP_ID is not set",
		})
		return
	}

	signingKey, err := GetPublicSigningKey(id)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	log.Println(id)
	log.Println(signingKey)
	c.JSON(200, signingKey)
}

func main() {
	r := gin.Default()
	r.GET("/.well-known/terraform.json", wellknown)
	r.GET("/v1/providers/:namespace/:name/versions", listVersions)
	r.GET("/v1/providers/:namespace/:name/:version/download/:os/:archi", download)
	r.POST("/v1/providers/regist", regist)
	r.Run()
}
