package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sqweek/dialog"
	"google.golang.org/api/option"

	drive "google.golang.org/api/drive/v3"
)

const (
	ServiceAccount = "CredentialFile.json" // Arquivo de credenciais
	filename       = ""                    // Nome do arquivo
	SCOPE          = drive.DriveScope
)

func main() {

MENU:
	dialog.Message("%s", "Escolha o arquivo").Title("Upload de Arquivos").Info()
	filename, err := dialog.File().Filter("", "").Load()
	fmt.Println(filename)

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(ServiceAccount), option.WithScopes(SCOPE))
	if err != nil {
		log.Fatalf("Não foi possivel criar serviço de Cliente %v", err)
	}

	file, err := os.Open(filename)
	info, _ := file.Stat()
	if err != nil {
		log.Fatalf("Nenhum arquivo selecionado, programa será encerrado: %v", err)
	}

	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	//Metadados
	f := &drive.File{
		Name:    info.Name(),
		Parents: []string{"YouFolderSortPath"}, //Caminho da pasta compartilhada
	}

	//Criando upload do arquivo
	res, err := srv.Files.
		Create(f).
		Media(file).
		ProgressUpdater(func(now, size int64) { fmt.Printf("%d, %d\r", now, size) }).
		Do()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Upload de arquivo com sucesso ID: %s\n", res.Id)
	if dialog.Message("%s", "Deseja selecionar outro Arquivo ?").Title("Novo Upload").YesNo() {
		goto MENU
	}
}
