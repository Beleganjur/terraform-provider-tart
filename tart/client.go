package tart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

type vmCreateRequest struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type vmCreateResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type vmResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

func apiURLJoin(base string, p string) string {
	// base like http://host:8085/api, p like /vms
	return fmt.Sprintf("%s%s", base, p)
}

func createVM(conf *config, name, image string) (string, string, error) {
	body, _ := json.Marshal(vmCreateRequest{Name: name, Image: image})
	req, err := http.NewRequest(http.MethodPost, apiURLJoin(conf.ApiURL, "/vms"), bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	if conf.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+conf.ApiToken)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var parsed vmCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", "", err
	}
	if parsed.ID == "" {
		parsed.ID = name
	}
	if parsed.Status == "" {
		parsed.Status = "running"
	}
	return parsed.ID, parsed.Status, nil
}

func getVM(conf *config, id string) (string, string, string, error) {
	url := apiURLJoin(conf.ApiURL, path.Join("/vms", id))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", "", err
	}
	if conf.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+conf.ApiToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", "", "", fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var parsed vmResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", "", "", err
	}
	return parsed.Name, parsed.Image, parsed.Status, nil
}

func deleteVM(conf *config, id string) error {
	url := apiURLJoin(conf.ApiURL, path.Join("/vms", id))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if conf.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+conf.ApiToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}
