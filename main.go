package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sqweek/dialog"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	ServiceAccount = "" // Arquivo de credenciais
	SCOPE          = drive.DriveScope
	FolderID       = "" // ID da pasta no Google Drive
)

func main() {
	// Exibir menu para escolher entre Upload, Excluir Arquivo, Listar Arquivos, Download ou Sair
	for {
		// Exibir opções para o usuário
		fmt.Println("\nEscolha uma opção:")
		fmt.Println("1. Efetuar Upload")
		fmt.Println("2. Excluir arquivo")
		fmt.Println("3. Listar arquivos")
		fmt.Println("4. Baixar arquivo")
		fmt.Println("5. Sair")

		var choice int
		fmt.Print("Digite sua opção (1, 2, 3, 4 ou 5): ")

		// Limpar o buffer de entrada
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Entrada inválida, tente novamente.")
			continue
		}

		switch choice {
		case 1:
			// Efetuar upload
			err := uploadFile()
			if err != nil {
				log.Fatalf("Erro no upload: %v", err)
			}
		case 2:
			// Excluir arquivo
			err := deleteFileByID()
			if err != nil {
				log.Fatalf("Erro ao excluir o arquivo: %v", err)
			}
		case 3:
			// Listar arquivos
			err := listFiles()
			if err != nil {
				log.Fatalf("Erro ao listar arquivos: %v", err)
			}
		case 4:
			// Baixar arquivo
			err := downloadFile()
			if err != nil {
				log.Fatalf("Erro ao baixar o arquivo: %v", err)
			}
		case 5:
			// Sair do programa
			fmt.Println("\nSaindo do programa.")
			return
		default:
			// Caso a opção seja inválida
			fmt.Println("Opção inválida, por favor escolha entre 1, 2, 3, 4 ou 5.")
		}
	}
}

// Função para fazer upload de um arquivo
func uploadFile() error {
	// Abrir o diálogo para escolher o arquivo para upload
	filename, err := dialog.File().Title("Escolha o arquivo para upload").Filter("Todos os Arquivos", "*.*").Load()
	if err != nil || filename == "" {
		return fmt.Errorf("erro ao escolher o arquivo para upload: %v", err)
	}

	// Verificar se o arquivo existe
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("o arquivo %s não existe", filename)
	}
	fmt.Println("\nArquivo selecionado:", filename)

	// Configuração do contexto e do serviço Google Drive
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(ServiceAccount), option.WithScopes(SCOPE))
	if err != nil {
		return fmt.Errorf("não foi possível criar o serviço de Cliente: %v", err)
	}

	// Abrir o arquivo para upload
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %v", err)
	}
	defer file.Close()

	// Obter informações do arquivo
	info, _ := file.Stat()

	// Metadados do arquivo (incluindo a pasta de destino)
	f := &drive.File{
		Name:    info.Name(),
		Parents: []string{FolderID}, // ID da pasta de destino
	}

	// Criar o upload do arquivo
	res, err := srv.Files.Create(f).Media(file).ProgressUpdater(func(now, size int64) {
		fmt.Printf("%d/%d bytes\r", now, size)
	}).Do()
	if err != nil {
		return fmt.Errorf("erro no upload: %v", err)
	}

	fmt.Printf("\nUpload de arquivo com sucesso! ID: %s\n", res.Id)
	return nil
}

// Função para excluir um arquivo por ID
func deleteFileByID() error {
	// Configuração do contexto e do serviço Google Drive
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(ServiceAccount), option.WithScopes(SCOPE))
	if err != nil {
		return fmt.Errorf("não foi possível criar o serviço de Cliente: %v", err)
	}

	// Listar arquivos na pasta
	files, err := listFilesInFolder(srv, FolderID)
	if err != nil {
		return fmt.Errorf("erro ao listar arquivos: %v", err)
	}

	// Se não houver arquivos, mostrar mensagem e retornar
	if len(files) == 0 {
		fmt.Println("\nNão há arquivos na pasta para excluir.")
		return nil
	}

	// Exibir os arquivos para o usuário escolher
	fmt.Println("\nArquivos encontrados na pasta:")
	for i, file := range files {
		fmt.Printf("%d. Nome: %s\n", i+1, file.Name)
		fmt.Printf("   ID: %s\n", file.Id)
		fmt.Println("--------------------------------------------")
	}

	// Perguntar ao usuário qual arquivo deseja excluir
	var choice int
	for {
		fmt.Print("\nEscolha o número do arquivo que deseja excluir: ")

		// Limpar o buffer de entrada
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Entrada inválida, tente novamente.")
			continue
		}

		// Verificar se o número escolhido está dentro do intervalo
		if choice < 1 || choice > len(files) {
			fmt.Println("Opção inválida, escolha um número entre 1 e", len(files))
			continue
		}
		break
	}

	// Obter o ID do arquivo escolhido
	fileID := files[choice-1].Id
	err = srv.Files.Delete(fileID).Do() // Aqui era necessário usar srv.Files.Delete
	if err != nil {
		return fmt.Errorf("erro ao excluir o arquivo: %v", err)
	}

	fmt.Println("\nArquivo excluído com sucesso!")
	return nil
}

// Função para listar arquivos na pasta especificada
func listFilesInFolder(srv *drive.Service, folderID string) ([]*drive.File, error) {
	// Listar arquivos dentro da pasta
	query := fmt.Sprintf("'%s' in parents", folderID)
	fileList, err := srv.Files.List().
		Q(query).
		Fields("files(id, name, mimeType, size)").Do()

	if err != nil {
		return nil, fmt.Errorf("erro ao listar arquivos: %v", err)
	}

	return fileList.Files, nil
}

// Função para listar arquivos com detalhes (tipo, extensão, tamanho)
func listFiles() error {
	// Configuração do contexto e do serviço Google Drive
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(ServiceAccount), option.WithScopes(SCOPE))
	if err != nil {
		return fmt.Errorf("não foi possível criar o serviço de Cliente: %v", err)
	}

	// Listar arquivos na pasta
	files, err := listFilesInFolder(srv, FolderID)
	if err != nil {
		return fmt.Errorf("erro ao listar arquivos: %v", err)
	}

	// Se não houver arquivos, mostrar mensagem e retornar
	if len(files) == 0 {
		fmt.Println("\nNão há arquivos na pasta.")
		return nil
	}

	// Exibir os arquivos para o usuário com detalhes (nome, tipo, extensão e tamanho)
	fmt.Println("\nArquivos encontrados na pasta:")
	for i, file := range files {
		// Obter tipo MIME e extensão do arquivo
		ext := filepath.Ext(file.Name)
		size := file.Size

		// Exibir as informações de maneira mais organizada
		fmt.Printf("\n%d. Nome: %s\n", i+1, file.Name)
		fmt.Printf("   Tipo MIME: %s\n", file.MimeType)
		fmt.Printf("   Extensão: %s\n", ext)
		if size > 0 {
			fmt.Printf("   Tamanho: %d bytes\n", size)
		} else {
			fmt.Println("   Tamanho: Não disponível")
		}
		fmt.Println("--------------------------------------------")
	}

	return nil
}

// Função para baixar um arquivo do Google Drive
func downloadFile() error {
	// Configuração do contexto e do serviço Google Drive
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(ServiceAccount), option.WithScopes(SCOPE))
	if err != nil {
		return fmt.Errorf("não foi possível criar o serviço de Cliente: %v", err)
	}

	// Listar arquivos na pasta
	files, err := listFilesInFolder(srv, FolderID)
	if err != nil {
		return fmt.Errorf("Erro ao listar arquivos: %v", err)
	}

	// Se não houver arquivos, mostrar mensagem e retornar
	if len(files) == 0 {
		fmt.Println("\nNão há arquivos na pasta para baixar.")
		return nil
	}

	// Exibir os arquivos para o usuário escolher
	fmt.Println("\nArquivos encontrados na pasta:")
	for i, file := range files {
		fmt.Printf("%d. Nome: %s (ID: %s)\n", i+1, file.Name, file.Id)
	}

	// Perguntar ao usuário qual arquivo deseja baixar
	var choice int
	for {
		fmt.Print("\nEscolha o número do arquivo que deseja baixar: ")

		// Limpar o buffer de entrada
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Entrada inválida, tente novamente.")
			continue
		}

		// Verificar se o número escolhido está dentro do intervalo
		if choice < 1 || choice > len(files) {
			fmt.Println("Opção inválida, escolha um número entre 1 e", len(files))
			continue
		}
		break
	}

	// Obter o arquivo escolhido
	fileID := files[choice-1].Id

	// Abrir o diálogo para escolher onde salvar o arquivo
	// Abrir o diálogo para escolher onde salvar o arquivo
	savePath, err := dialog.File().Title("Escolha onde salvar o arquivo").Filter("Todos os Arquivos", "*.*").Save()
	if err != nil || savePath == "" {
		return fmt.Errorf("erro ao escolher o local para salvar o arquivo: %v", err)
	}

	// Criar o arquivo local onde o conteúdo será salvo
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo para salvar: %v", err)
	}
	defer file.Close()

	// Baixar o conteúdo do arquivo do Google Drive
	res, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return fmt.Errorf("Erro ao baixar o arquivo: %v", err)
	}
	defer res.Body.Close()

	// Copiar o conteúdo para o arquivo local
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return fmt.Errorf("Erro ao salvar o arquivo: %v", err)
	}

	fmt.Printf("\nArquivo baixado com sucesso para: %s\n", savePath)
	return nil
}
