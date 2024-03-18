package main

import (
	"context"
	"encoding/json"
	"jcamilofarfan/admin_github_bot/src/utils"
	"os"
	"reflect"

	"github.com/google/go-github/v60/github"
)

type MyRepository struct {
	Id           *int64   `json:"id"`
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	Homepage     *string  `json:"homepage"`
	AllowForking *bool    `json:"allow_forking"`
	Topics       []string `json:"topics"`
	Archived     *bool    `json:"archived"`
	Disabled     *bool    `json:"disabled"`
	Private      *bool    `json:"private"`
	IsTemplate   *bool    `json:"is_template"`
}

var client *github.Client
var owner string
var update_json bool = false

func main() {
	utils.Log(utils.Info, "Iniciando el bot")
	token := utils.GetEnv("GITHUB_TOKEN", "")
	if token == "" {
		utils.Log(utils.Error, "No se ha encontrado el token de autenticación")
	}
	utils.Log(utils.Info, "Token de autenticación encontrado")
	client = github.NewClient(nil).WithAuthToken(token)
	utils.Log(utils.Info, "Cliente de github creado")
	client.UserAgent = "admin_github_bot"
	utils.Log(utils.Info, "UserAgent creado")
	user, reponse, err := client.Users.Get(context.Background(), "")
	utils.Log(utils.Info, "Status: %v", reponse.Status)
	if err != nil {
		utils.Log(utils.Error, "Error al obtener el usuario %v", err)
	}
	owner = *user.Login
	my_repositories_github := get_repositories()
	compare_repositories(my_repositories_github)
	if update_json {
		my_repositories_github = get_repositories()
		create_json(my_repositories_github)
	}
}

func get_repositories() []MyRepository {
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}
	var all_repos []*github.Repository
	for {
		list_repo, resp, err_getting_repos := client.Repositories.ListByAuthenticatedUser(context.Background(), opt)
		if err_getting_repos != nil {
			utils.Log(utils.Error, "Error al obtener los repositorios %v", err_getting_repos)
		}
		all_repos = append(all_repos, list_repo...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

	}
	utils.Log(utils.Info, "Listado de repositorios obtenido")
	my_repositories_github := parse_to_my_repository(all_repos)
	utils.Log(utils.Info, "Repositorios parseados")
	return my_repositories_github
}

func create_json(list_repo []MyRepository) {
	utils.Log(utils.Info, "Inicio de creación de json")
	result_json, err_marshal := json.Marshal(list_repo)
	if err_marshal != nil {
		utils.Log(utils.Error, "Error al crear el json")
	}
	file, err_create_file := os.Create("jcamilofarfan.json")
	if err_create_file != nil {
		utils.Log(utils.Error, "Error al crear el archivo")
	}
	defer file.Close()
	_, err_write_file := file.Write(result_json)
	if err_write_file != nil {
		utils.Log(utils.Error, "Error al escribir el archivo")
	}
	utils.Log(utils.Info, "Fin de creación de json")
}

func parse_to_my_repository(list_repo []*github.Repository) []MyRepository {
	utils.Log(utils.Info, "Inicio de parseo de repositorio")
	var my_repositories []MyRepository
	for _, repo := range list_repo {
		my_repo_json, err_marshal := json.Marshal(repo)
		if err_marshal != nil {
			utils.Log(utils.Error, "Error al crear el json")
		}
		my_repo := NewMyRepository()
		err_unmarshal := json.Unmarshal(my_repo_json, &my_repo)
		if err_unmarshal != nil {
			utils.Log(utils.Error, "Error al decodificar el json")
		}
		my_repositories = append(my_repositories, my_repo)
	}
	utils.Log(utils.Info, "Fin de parseo de repositorio")
	return my_repositories
}

func NewMyRepository() MyRepository {
	return MyRepository{
		Id:           new(int64),  // 0
		Name:         new(string), // ""
		Description:  new(string), // ""
		Homepage:     new(string), // ""
		AllowForking: new(bool),   // false
		Topics:       make([]string, 0),
		Archived:     new(bool), // false
		Disabled:     new(bool), // false
		Private:      new(bool), // false
		IsTemplate:   new(bool), // false
	}
}

func compare_repositories(list_repo_github []MyRepository) {
	utils.Log(utils.Info, "Inicio de comparación de repositorios")
	list_repo_local := get_json_file()
	github_slice := make([]interface{}, len(list_repo_github))
	local_slice := make([]interface{}, len(list_repo_local))
	for i, v := range list_repo_github {
		github_slice[i] = v
	}
	for i, v := range list_repo_local {
		local_slice[i] = v
	}
	for _, repo := range github_slice {
		repo_local := find(local_slice, func(p interface{}) bool {
			return *p.(MyRepository).Id == *repo.(MyRepository).Id
		})
		if repo_local == nil {
			utils.Log(utils.Info, "Se debe eliminar el repositorio: %s", *repo.(MyRepository).Name)
			eliminar_repo(repo.(MyRepository))
		}
	}
	for _, repo := range local_slice {
		repo_github := find(github_slice, func(p interface{}) bool {
			return *p.(MyRepository).Id == *repo.(MyRepository).Id
		})
		if repo_github == nil {
			utils.Log(utils.Info, "Se debe crear el repositorio: %s", *repo.(MyRepository).Name)
			crear_repo(repo.(MyRepository))
			continue
		}
		repo_github_struct := repo_github.(MyRepository)
		compare_repository(repo.(MyRepository), repo_github_struct)
	}

	utils.Log(utils.Info, "Fin de comparación de repositorios")
}

func eliminar_repo(repo MyRepository) {
	utils.Log(utils.Info, "Inicio de eliminación de repositorio")
	response, err := client.Repositories.Delete(context.Background(), owner, *repo.Name)
	utils.Log(utils.Info, "Status: %v", response.Status)
	if err != nil {
		utils.Log(utils.Error, "Error al eliminar el repositorio %v", err)
	}
	utils.Log(utils.Info, "Fin de eliminación de repositorio")
}

func actualizar_repo(repo github.Repository, repo_name string) {
	utils.Log(utils.Info, "Inicio de actualización de repositorio %v", repo_name)
	_, response, err := client.Repositories.Edit(context.Background(), owner, repo_name, &repo)
	utils.Log(utils.Info, "Status: %v", response.Status)
	if err != nil {
		utils.Log(utils.Error, "Error al eliminar el repositorio %v", err)
	}
	utils.Log(utils.Info, "Fin de actualización de repositorio")
}

func crear_repo(repo MyRepository) {
	utils.Log(utils.Info, "Inicio de creación de repositorio")
	repository := &github.Repository{
		Name:         repo.Name,
		Description:  repo.Description,
		Homepage:     repo.Homepage,
		AllowForking: repo.AllowForking,
		Topics:       repo.Topics,
		Archived:     repo.Archived,
		Disabled:     repo.Disabled,
		Private:      repo.Private,
		IsTemplate:   repo.IsTemplate,
	}
	_, response, err := client.Repositories.Create(context.Background(), "", repository)
	utils.Log(utils.Info, "Status: %v", response.Status)
	if err != nil {
		utils.Log(utils.Error, "Error al crear el repositorio %v", err)
	}
	update_json = true
	utils.Log(utils.Info, "Fin de creación de repositorio")
}

func find(array []interface{}, condicion func(interface{}) bool) interface{} {
	for _, p := range array {
		if condicion(p) {
			return p
		}
	}
	return nil
}

func get_json_file() []MyRepository {
	utils.Log(utils.Info, "Inicio de obtención de json")
	response := []MyRepository{}
	file, err_open_file := os.Open("jcamilofarfan.json")
	if err_open_file != nil {
		utils.Log(utils.Error, "Error al abrir el archivo")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err_decode := decoder.Decode(&response)
	if err_decode != nil {
		utils.Log(utils.Error, "Error al decodificar el archivo")
	}
	utils.Log(utils.Info, "Fin de obtención de json")
	return response
}

func compare_repository(local MyRepository, github_repo MyRepository) {
	update_repository := false
	repository_github_updated := &github.Repository{}
	if *local.Name != *github_repo.Name {
		repository_github_updated.Name = local.Name
		update_repository = true
	}
	if *local.Description != *github_repo.Description {
		repository_github_updated.Description = local.Description
		update_repository = true
	}
	if *local.Homepage != *github_repo.Homepage {
		repository_github_updated.Homepage = local.Homepage
		update_repository = true
	}
	if *local.AllowForking != *github_repo.AllowForking {
		repository_github_updated.AllowForking = local.AllowForking
		update_repository = true
	}
	if !reflect.DeepEqual(local.Topics, github_repo.Topics) {
		repository_github_updated.Topics = local.Topics
		update_repository = true
	}
	if *local.Archived != *github_repo.Archived {
		repository_github_updated.Archived = local.Archived
		update_repository = true
	}
	if *local.Disabled != *github_repo.Disabled {
		repository_github_updated.Disabled = local.Disabled
		update_repository = true
	}
	if *local.Private != *github_repo.Private {
		repository_github_updated.Private = local.Private
		update_repository = true
	}
	if *local.IsTemplate != *github_repo.IsTemplate {
		repository_github_updated.IsTemplate = local.IsTemplate
		update_repository = true
	}
	if update_repository {
		actualizar_repo(*repository_github_updated, *github_repo.Name)
	}
}
