package main

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func createToken(cfg Config) (string, error) {
	headers := map[string]interface{}{
		"alg": "ES256",
		"kid": cfg.Kid,
		"typ": "JWT",
	}

	now := time.Now()
	expirationTime := now.Add(19 * time.Minute)
	block, _ := pem.Decode([]byte(cfg.P8Key))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	signingKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	//bb, _ := json.Marshal(signingKey.(*ecdsa.PrivateKey).PublicKey)
	//fmt.Println(base64.StdEncoding.EncodeToString(bb))

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": cfg.Iis,
		"iat": now.Unix(),
		"exp": expirationTime.Unix(),
		"aud": "appstoreconnect-v1",
	})
	token.Header = headers

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	// fmt.Printf("jwt token: %s\n", signedToken)
	return signedToken, nil
}

const (
	GREEN = "\033[0;32m"
	RED   = "\033[0;31m"
	GRAY  = "\033[1;30m"
	NC    = "\033[0m" // No Color

)

func main() {
	reader := bufio.NewReader(os.Stdin)

	configFile, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	cfg := Config{}
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(GREEN + ">> генерируем токен для App Store Connect" + NC)
	jwtToken, err := createToken(cfg)
	if err != nil {
		panic(err)
	}

	head := http.Header{
		"Authorization": []string{"Bearer " + jwtToken},
	}

	fmt.Println(GREEN + "\n>> получаем список всех доступных приложении в App Store Connect" + NC)
	var (
		appsInfo Apps
		url      = "https://api.appstoreconnect.apple.com/v1/apps"
	)

	fmt.Printf("%s[debug] этап %d. ссылка %s%s\n", GRAY, 1, url, NC)
	err = doRequest(context.Background(), "GET", url, head, &appsInfo)
	if err != nil {
		panic(err)
	}

	if len(appsInfo.Data) == 0 {
		err = fmt.Errorf("%sприложения не найдены!%s", RED, NC)
		panic(err)
	}

	fmt.Printf("\n%sнайденные приложения:%s\n", GRAY, NC)
	for i := range appsInfo.Data {
		fmt.Printf("%s%d. %s (%s)%s\n", GREEN, i+1, appsInfo.Data[i].Attributes.Name, appsInfo.Data[i].Attributes.BundleID, NC)
	}

	var appDir string
	for {
		fmt.Printf("%s %s %s", RED, "\nвыберите приложение, введите номер\n: ", NC)
		number, _ := reader.ReadString('\n')
		number = strings.Replace(number, "\n", "", -1)

		n, _ := strconv.Atoi(number)
		url = appsInfo.Data[n-1].Relationships.CiProduct.Links.Related
		appDir = "./" + appsInfo.Data[n-1].Attributes.BundleID

		fmt.Printf("\n%s>> выбран нмоер %d: %s%s\n", GREEN, n, appsInfo.Data[n-1].Attributes.BundleID, NC)
		break
	}

	if url == "" {
		panic("ссылка `ciProduct` не найдена")
	}

	var ciProduct CiProduct
	fmt.Printf("%s\n[debug] этап %d. ссылка %s%s\n", GRAY, 2, url, NC)
	err = doRequest(context.Background(), "GET", url, head, &ciProduct)
	if err != nil {
		panic(err)
	}

	url = ciProduct.Data.Relationships.BuildRuns.Links.Related
	if url == "" {
		panic("ссылка `buildRuns` не найдена")
	}

	var buildRuns BuildRuns
	url = url + "?limit=200"
	fmt.Printf("%s[debug] этап %d. ссылка %s%s\n", GRAY, 3, url, NC)
	err = doRequest(context.Background(), "GET", url, head, &buildRuns)
	if err != nil {
		panic(err)
	}

	lastBuildId := 0
	url = ""
	for i := range buildRuns.Data {
		if buildRuns.Data[i].Attributes.ExecutionProgress == "COMPLETE" && buildRuns.Data[i].Attributes.CompletionStatus == "SUCCEEDED" {
			lastBuildId = buildRuns.Data[i].Attributes.Number
			url = buildRuns.Data[i].Relationships.Actions.Links.Related
		}
	}
	fmt.Printf("%s[debug] этап %d. номер сборки: %s%d%s\n", GRAY, 3, GREEN, lastBuildId, NC)

	if url == "" {
		panic("ссылка `actions` не найдена")
	}
	var actions Actions
	fmt.Printf("%s[debug] этап %d. ссылка %s%s\n", GRAY, 4, url, NC)
	err = doRequest(context.Background(), "GET", url, head, &actions)
	if err != nil {
		panic(err)
	}

	url = ""
	for i := range actions.Data {
		if actions.Data[i].Attributes.ActionType == "ARCHIVE" && actions.Data[i].Attributes.ExecutionProgress == "COMPLETE" && actions.Data[i].Attributes.CompletionStatus == "SUCCEEDED" {
			url = actions.Data[i].Relationships.Artifacts.Links.Related
		}
	}

	if url == "" {
		panic("ссылка `artifacts` не найдена")
	}
	var ciartifacts CIArtifacts
	fmt.Printf("%s[debug] этап %d. ссылка %s%s\n", GRAY, 4, url, NC)
	err = doRequest(context.Background(), "GET", url, head, &ciartifacts)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(appDir, 0777)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	for i := range ciartifacts.Data {
		var cifiles DownloadURL

		url = ciartifacts.Data[i].Links.Self
		fmt.Printf("%s[debug] этап %d. ссылка %s%s\n", GRAY, 5, url, NC)
		err = doRequest(context.Background(), "GET", url, head, &cifiles)
		if err != nil {
			panic(err)
		}

		if url == "" {
			panic("ссылка `downloadUrl` не найдена")
		}
		fmt.Printf("%s[debug] этап %d. тип %s. имя %s%s\n", GRAY, 5, cifiles.Data.Attributes.FileType, cifiles.Data.Attributes.FileName, NC)
		err = fileSave(appDir, cifiles.Data.Attributes.FileType+"-"+cifiles.Data.Attributes.FileName, cifiles.Data.Attributes.DownloadURL, head)
		if err != nil {
			panic(err)
		}
	}
}

func fileSave(directory, name, url string, head http.Header) (err error) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	req.Header = head

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(directory+"/"+name, bodyBytes, 0666)
	if err != nil {
		return err
	}
	fmt.Printf("%s[debug] этап %d. ссылка на файл %s%s\n", GRAY, 5, url, NC)
	fmt.Printf("%s>> скачан файл %s%s\n", GREEN, name, NC)

	return
}

func doRequest(ctx context.Context, method, url string, request_header http.Header, response any) (err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return
	}

	for k, v := range request_header {
		req.Header.Set(k, v[0])
	}

	var resp *http.Response
	c := &http.Client{}
	resp, err = c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if resp.Body != http.NoBody {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			err = json.Unmarshal(bodyBytes, response)
		}
	} else {
		err = errors.New("response body is empty")
	}

	return
}

type Links struct {
	Self    string `json:"self"`
	Related string `json:"related"`
}

type Meta struct {
	Paging struct {
		Limit int `json:"limit"`
	} `json:"paging"`
}

type AppInfo struct {
	Data struct {
		Relationships struct {
			CiProduct struct {
				Links Links `json:"links"`
			} `json:"ciProduct"`
		} `json:"relationships"`
	} `json:"data"`
}

type CiProduct struct {
	Data struct {
		Relationships struct {
			BuildRuns struct {
				Links Links `json:"links"`
			} `json:"buildRuns"`
		} `json:"relationships"`
	} `json:"data"`
}

type BuildRuns struct {
	Data []struct {
		Attributes struct {
			Number            int    `json:"number"`
			FinishedDate      string `json:"finishedDate"`
			ExecutionProgress string `json:"executionProgress"`
			CompletionStatus  string `json:"completionStatus"`
		} `json:"attributes"`
		Relationships struct {
			Builds struct {
				Links Links `json:"links"`
			} `json:"builds"`
			Actions struct {
				Links Links `json:"links"`
			} `json:"actions"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type Actions struct {
	Data []struct {
		Attributes struct {
			ActionType        string    `json:"actionType"`
			ExecutionProgress string    `json:"executionProgress"`
			Name              string    `json:"name"`
			CompletionStatus  string    `json:"completionStatus"`
			FinishedDate      time.Time `json:"finishedDate"`
		} `json:"attributes"`
		Relationships struct {
			Artifacts struct {
				Links Links `json:"links"`
			} `json:"artifacts"`
		} `json:"relationships"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type Artifacts struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			ActionType        string    `json:"actionType"`
			IssueCounts       any       `json:"issueCounts"`
			ExecutionProgress string    `json:"executionProgress"`
			Name              string    `json:"name"`
			StartedDate       time.Time `json:"startedDate"`
			CompletionStatus  string    `json:"completionStatus"`
			IsRequiredToPass  bool      `json:"isRequiredToPass"`
			FinishedDate      time.Time `json:"finishedDate"`
		} `json:"attributes"`
		Relationships struct {
			Artifacts struct {
				Links Links `json:"links"`
			} `json:"artifacts"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type CIArtifacts struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			FileName string `json:"fileName"`
			FileSize int    `json:"fileSize"`
			FileType string `json:"fileType"`
		} `json:"attributes"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type DownloadURL struct {
	Data struct {
		Attributes struct {
			FileType    string `json:"fileType"`
			FileName    string `json:"fileName"`
			FileSize    int    `json:"fileSize"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"attributes"`
	} `json:"data"`
}

type Apps struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Name     string `json:"name"`
			BundleID string `json:"bundleId"`
		} `json:"attributes"`
		Relationships struct {
			CiProduct struct {
				Links Links `json:"links"`
			} `json:"ciProduct"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
}

type Config struct {
	Kid   string `json:"kid"`
	Iis   string `json:"iis"`
	P8Key string `json:"p8key"`
}
