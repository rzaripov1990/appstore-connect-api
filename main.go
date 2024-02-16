package main

import (
	appstore_jwt "appstore-connect-api/appstore-jwt"
	"appstore-connect-api/config"
	"appstore-connect-api/console"
	"appstore-connect-api/models"
	"appstore-connect-api/rpc"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
)

func main() {
	console.Print("%s", "\n🚀 App Store Connect API").NL()

	reader := console.New()
	cfg := config.New()

	console.PrintGray("%s", ">> генерируем токен для App Store Connect")
	jwtToken, err := appstore_jwt.CreateToken(cfg)
	if err != nil {
		console.PrintRed(" - %s", "FAIL").NL()
		console.PrintRed("[error] %s", err.Error()).NL()
		os.Exit(1)
	}
	console.PrintGreen(" - %s", "OK").NL()

	head := http.Header{
		"Authorization": []string{"Bearer " + jwtToken},
	}

	var (
		appsInfo models.Apps
		url      = "https://api.appstoreconnect.apple.com/v1/apps"
	)

	{

		console.PrintGray("%s", ">> получаем список всех доступных приложении в App Store Connect")
		err = rpc.Do(context.Background(), "GET", url, head, &appsInfo)
		if err != nil {
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}

		if len(appsInfo.Data) == 0 {
			err = errors.New("apps not found")
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
		}
		console.PrintGreen(" - %s", "OK").NL()
	}

	var appDir string
	{
		console.PrintGray("%s", ">> найденные приложения").NL()
		for i := range appsInfo.Data {
			console.PrintGreen("%d. %s (%s)", i+1, appsInfo.Data[i].Attributes.Name, appsInfo.Data[i].Attributes.BundleID).NL()
		}

		console.PrintGray("%s", "\n>> выберите приложение, введите номер\n: ")

		number := reader.ReadInput()
		n, _ := strconv.Atoi(number)

		url = appsInfo.Data[n-1].Relationships.CiProduct.Links.Related
		appDir = "./" + appsInfo.Data[n-1].Attributes.BundleID

		if url == "" {
			err = errors.New("ссылка `ciProduct` не найдена")
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
		}
		console.PrintGreen("\n>> выбран номер %d: %s", n, appsInfo.Data[n-1].Attributes.BundleID).NL()
	}

	var ciProduct models.CiProduct
	{
		console.PrintGray("[debug] этап %d. ссылка %s", 2, url)
		err = rpc.Do(context.Background(), "GET", url, head, &ciProduct)
		if err != nil {
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}

		url = ciProduct.Data.Relationships.BuildRuns.Links.Related
		if url == "" {
			err = errors.New("ссылка `buildRuns` не найдена")
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
		}
		console.PrintGreen(" - %s", "OK").NL()
	}

	var buildRuns models.BuildRuns
	{
		url = url + "?limit=200"
		console.PrintGray("[debug] этап %d. ссылка %s", 3, url)
		err = rpc.Do(context.Background(), "GET", url, head, &buildRuns)
		if err != nil {
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}

		lastBuildId := 0
		url = ""
		for i := range buildRuns.Data {
			if buildRuns.Data[i].Attributes.ExecutionProgress == "COMPLETE" && buildRuns.Data[i].Attributes.CompletionStatus == "SUCCEEDED" {
				lastBuildId = buildRuns.Data[i].Attributes.Number
				url = buildRuns.Data[i].Relationships.Actions.Links.Related
			}
		}

		if url == "" {
			err = errors.New("ссылка `actions` не найдена")
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
		}

		console.PrintGreen(" - %s", "OK").NL()
		console.PrintGreen(">> последняя сборка: %d", lastBuildId).NL()
	}

	var actions models.Actions
	{
		console.PrintGray("[debug] этап %d. ссылка %s", 4, url)
		err = rpc.Do(context.Background(), "GET", url, head, &actions)
		if err != nil {
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}

		url = ""
		for i := range actions.Data {
			if actions.Data[i].Attributes.ActionType == "ARCHIVE" && actions.Data[i].Attributes.ExecutionProgress == "COMPLETE" && actions.Data[i].Attributes.CompletionStatus == "SUCCEEDED" {
				url = actions.Data[i].Relationships.Artifacts.Links.Related
			}
		}

		if url == "" {
			err = errors.New("ссылка `artifacts` не найдена")
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
		}
		console.PrintGreen(" - %s", "OK").NL()
	}

	var ciartifacts models.CIArtifacts
	{
		console.PrintGray("[debug] этап %d. ссылка %s", 5, url)
		err = rpc.Do(context.Background(), "GET", url, head, &ciartifacts)
		if err != nil {
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}
		console.PrintGreen(" - %s", "OK").NL()

	createDir:
		console.PrintGreen(">> создание директории %s", appDir)
		err = os.Mkdir(appDir, 0777)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintGreen("%s", ">> удалить папку? (y/n)\n: ")

				answer := reader.ReadInput()
				if answer == "y" {
					os.RemoveAll(appDir)
					goto createDir
				}
			}
			console.PrintGreen(">> создание директории %s", appDir)
			console.PrintRed(" - %s", "FAIL").NL()
			console.PrintRed("[error] %s", err.Error()).NL()
			os.Exit(1)
		}
		console.PrintGreen(" - %s", "OK").NL()

		for i := range ciartifacts.Data {
			var cifiles models.DownloadURL

			url = ciartifacts.Data[i].Links.Self
			console.PrintGray("[debug] этап %d. ссылка %s", 5, url)
			err = rpc.Do(context.Background(), "GET", url, head, &cifiles)
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}

			if url == "" {
				err = errors.New("ссылка `downloadUrl` не найдена")
				if err != nil {
					console.PrintRed(" - %s", "FAIL").NL()
					console.PrintRed("[error] %s", err.Error()).NL()
					os.Exit(1)
				}
			}
			console.PrintGreen(" - %s", "OK").NL()

			console.PrintGray("[debug] этап %d. тип %s. имя %s", 5, cifiles.Data.Attributes.FileType, cifiles.Data.Attributes.FileName)
			err = rpc.FileSave(appDir, cifiles.Data.Attributes.FileType+"-"+cifiles.Data.Attributes.FileName, cifiles.Data.Attributes.DownloadURL, head)
			if err != nil {
				console.PrintRed(" - %s", "FAIL").NL()
				console.PrintRed("[error] %s", err.Error()).NL()
				os.Exit(1)
			}
			console.PrintGreen(" - %s", "OK").NL()
		}
	}
}
