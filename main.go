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

type Provider struct {
	Protocols          []string    `json:"protocols"`
	OS                 string      `json:"os"`
	Arch               string      `json:"arch"`
	Filename           string      `json:"filename"`
	DownloadURL        string      `json:"download_url"`
	ShasumsURL         string      `json:"shasums_url"`
	ShasumSignatureURL string      `json:"shasums_signature_url"`
	Shasum             string      `json:"shasum"`
	SigningKeys        SigningKeys `json:"signing_keys"`
}

type SigningKeys struct {
	GpgPublicKeys []GpgPublicKey `json:"gpg_public_keys"`
}

type GpgPublicKey struct {
	KeyID          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SorceURL       string `json:"source_url"`
}

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

	id := os.Getenv("PGP_ID")
	if id == "" {
		c.JSON(500, gin.H{
			"message": "PGP_ID is not set",
		})
		return
	}

	keyFile := os.Getenv("PGP_PUBLIC_SIGNING_KEY_FILE")
	if id == "" {
		c.JSON(500, gin.H{
			"message": "PGP_ID is not set",
		})
		return
	}

	// signingKeyPrivate, err := GetPublicSigningKey(id)
	signingKeyPrivate, err := GetPublicSigningKeyFromFile(id, keyFile)
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	var p Provider
	if err := c.ShouldBindJSON(&p); err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	signingKeys := SigningKeys{
		GpgPublicKeys: []GpgPublicKey{
			{
				KeyID:      signingKeyPrivate.KeyID,
				AsciiArmor: signingKeyPrivate.ASCIIArmor,
			},
		},
	}
	p.SigningKeys = signingKeys

	file, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = os.MkdirAll(fmt.Sprintf("provider/%s/%s/", ns, name), os.ModePerm)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = ioutil.WriteFile(
		fmt.Sprintf("provider/%s/%s/%s_%s_%s.json", ns, name, p.OS, p.Arch, version),
		file,
		0644,
	)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, "ok")
}

func main() {
	r := gin.Default()
	r.GET("/.well-known/terraform.json", wellknown)
	r.GET("/v1/providers/:namespace/:name/versions", listVersions)
	r.GET("/v1/providers/:namespace/:name/:version/download/:os/:archi", download)
	r.POST("/v1/providers/:namespace/:name/:version/regist", regist)
	r.Run()
}
