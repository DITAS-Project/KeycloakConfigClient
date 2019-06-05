/*
 * Copyright 2018 Information Systems Engineering, TU Berlin, Germany
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *                       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * This is being developed for the DITAS Project: https://www.ditas-project.eu/
 */

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/DITAS-Project/KeycloakConfigClient/kcc"
	"go.uber.org/zap"
	"gopkg.in/gookit/color.v1"
)

var logger *zap.Logger
var log *zap.SugaredLogger

func init() {
	lgr, _ := zap.NewProduction()
	logger = lgr
	log = logger.Sugar()
}

func main() {
	defer logger.Sync()

	address := flag.String("address", "", "KeyCloakConfig API address")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	unsecured := flag.Bool("unsecured", true, "trust all ssl certificates")
	headless := flag.Bool("headless", false, "run without cli")

	blueprintFile := flag.String("bl", "", "blueprint file to commit")
	configFile := flag.String("cf", "", "relam config to commit,needs to have blueprint id.")

	flag.Parse()

	if verbose != nil && *verbose {
		logger.WithOptions(zap.Development())
	}

	if unsecured != nil && *unsecured {
		log.Warn("running in unsecured mode use --unsecured to change this!")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var endpoint string
	if *headless && *address == "" {
		log.Panic("can't run in headless without address.")
	}

	if *address == "" {
		address, err := ReadString("What is the keycloak-config endpoint you want to use?", color.LightBlue, -1, nil)
		if err != nil {
			log.Panic("need address to function")
		}
		endpoint = *address
	} else {
		endpoint = *address
	}

	client, err := NewKCC(endpoint)

	if err != nil {
		log.Panic("failed to create client", err)
	}

	var blueprint = BluePrint{}
	var config = Config{}

	if *headless {
		if *blueprintFile != "" {
			data, err := ioutil.ReadFile(*blueprintFile)
			if err != nil {
				log.Panic("can't read file", *blueprintFile, err)
			}
			err = json.Unmarshal(data, &blueprint)
			if err != nil {
				log.Panic("can't read file", *blueprintFile, err)
			}
			err = client.SendBlueprint(blueprint)
			if err != nil{
				log.Error(err)
			}
		}

		if *configFile != "" {
			data, err := ioutil.ReadFile(*configFile)
			if err != nil {
				log.Panic("can't read file", *configFile, err)
			}
			err = json.Unmarshal(data, &config)
			if err != nil {
				log.Panic("can't read file", *configFile, err)
			}
			err = client.SendConfig(config)
			if err != nil{
				log.Error(err)
			}
		}
		os.Exit(0)
	} else {
		for {
			num := Menu("What do you want to do?", color.LightGreen, []string{
				"Create a new Blueprint Realm",
				"Create or Update a Realm Config",
				"quit",
			})

			switch num {
			case 0:
				blueprintCommand(blueprint, client)
			case 1:
				configCommand(blueprint, config, client)
			case 2:
				{
					color.LightGreen.Println("Bye.")
					os.Exit(0)
				}
			}
		}
	}
}

func configCommand(blueprint BluePrint, config Config, client *ConfigClient) {
	doConfig := SimpleQuestion("Do you want to create/update a Config", color.LightBlue)
	if doConfig {
		if blueprint.BlueprintID == "" {
			bid, err := ReadString("What is the BlueprintID?", color.LightBlue, -1, nil)
			blueprint.BlueprintID = *bid
			if err != nil {
				log.Panic("Need the BlueprintID to create a config!")
			}
		}

		fileLoad := SimpleQuestion("Load a User Config form a file?", color.LightBlue)
		if fileLoad {
			path, err := ReadString("Enter the path to the file you want to load:", color.LightBlue, -1, func(path string) bool {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return false
				}
				return true
			})
			if err != nil {
				log.Panic("can't load this file")
			}
			data, err := ioutil.ReadFile(*path)
			if err != nil {
				log.Panic("can't read file", *path, err)
			}
			config.BlueprintID = blueprint.BlueprintID

			err = json.Unmarshal(data, &config)
			if err != nil {
				log.Panic("can't read file", *path, err)
			}
		} else {
			config.BlueprintID = blueprint.BlueprintID

			roles := make([]string, 0)
			for {
				role, _ := ReadString("Enter a role name used of your VDC", color.LightBlue, -1, nil)
				roles = append(roles, *role)
				ctn := SimpleQuestion("Add another role?", color.LightBlue)
				if !ctn {
					break
				}
			}

			config.Roles = roles

			roleSelect := ""
			for i, role := range roles {
				roleSelect += fmt.Sprintf("\t%d %s\n", i, role)
			}

			users := make([]UserConfig, 0)
			for {
				user := UserConfig{}
				name, _ := ReadString("Enter a username ", color.LightBlue, -1, nil)
				pwd, _ := ReadString("Enter a password ", color.LightBlue, -1, nil)

				user.Username = *name
				user.Password = *pwd
				user.Roles = make([]string, 0)
				for {
					i := Menu(fmt.Sprintf("Select role for %s", *name), color.LightBlue, user.Roles)
					user.Roles = append(user.Roles, roles[i])

					ctn := SimpleQuestion("Add another role?", color.LightBlue)
					if !ctn {
						break
					}
				}
				users = append(users, user)

				ctn := SimpleQuestion("Add another user?", color.LightBlue)
				if !ctn {
					break
				}
			}
			config.Users = users

		}
		apply := SimpleQuestion(fmt.Sprintf("Do you want to commit this config?\n%+v", config), color.LightMagenta)
		if apply {
			err := client.SendConfig(config)
			if err != nil {
				log.Panic("Need the BlueprintID to create a config!")
			}
		}
	}
}

func blueprintCommand(blueprint BluePrint, client *ConfigClient) {
	createBlueprint := SimpleQuestion("Do you want to initialized a new Blueprint?", color.LightBlue)
	if createBlueprint {
		fileLoad := SimpleQuestion("Load a Blueprint Config form a File?", color.LightBlue)

		if fileLoad {
			path, err := ReadString("Enter the path to the file you want to load:", color.LightBlue, -1, func(path string) bool {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return false
				}
				return true
			})
			if err != nil {
				log.Panic("can't load this file")
			}
			data, err := ioutil.ReadFile(*path)
			if err != nil {
				log.Panic("can't read file", *path, err)
			}
			err = json.Unmarshal(data, &blueprint)
			if err != nil {
				log.Panic("can't read file", *path, err)
			}
		} else {
			bid, _ := ReadString("What is the BlueprintID?", color.LightBlue, -1, nil)
			blueprint.BlueprintID = *bid

			cid, _ := ReadString("What is the clientID?", color.LightBlue, -1, nil)
			blueprint.ClientId = *cid
		}

		apply := SimpleQuestion(fmt.Sprintf("Do you want to commit this config?\n%+v", blueprint), color.LightMagenta)
		if apply {
			//TODO: print output!
			err := client.SendBlueprint(blueprint)
			if err != nil {
				log.Debug("failed to send blueprint")
				//TODO:
			}
		}

	}
}
