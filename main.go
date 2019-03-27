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
	"go.uber.org/zap"
	"gopkg.in/gookit/color.v1"
	"io/ioutil"
	"net/http"
	"os"
)

var logger *zap.Logger
var log *zap.SugaredLogger

func init(){
	lgr, _ := zap.NewProduction()
	logger = lgr
	log = logger.Sugar()
}

func main() {
	defer logger.Sync()

	address := flag.String("address","","KeyCloakConfig API address")
	verbose := flag.Bool("verbose",false,"Enable verbose logging")
	unsecured := flag.Bool("unsecured",true,"trust all ssl certificates")

	flag.Parse()

	if verbose != nil && *verbose {
		logger.WithOptions(zap.Development())
	}

	if unsecured != nil && *unsecured {
		log.Warn("running in unsecured mode use --unsecured to change this!")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var endpoint string
	if *address == ""{
		address, err := readString("What is the keycloak-config endpoint you want to use?", color.LightBlue, -1, nil)
		if err != nil{
			log.Panic("need address to function")
		}
		endpoint = *address
	} else {
		endpoint = *address
	}

	client,err := NewKCC(endpoint)

	if err != nil{
		log.Error("failed to create client",err)
		color.Red.Println("Could not create client.")
	}
	var blueprint = BluePrint{}
	var config = Config{}
	for {
		num := menu("What do you want to do?",color.LightGreen,[]string{
			"Create a new Blueprint Realm",
			"Create or Update a Realm Config",
			"quit",
		})

		switch num {
			case 0:
				blueprintCommand(blueprint, client)
			case 1:
				configCommand(blueprint, config, client)
			case 2: {
				color.LightGreen.Println("Bye.")
				os.Exit(0)
			}
		}
	}
}

func configCommand(blueprint BluePrint, config Config, client *ConfigClient) {
	doConfig := simpleQuestion("Do you want to create/update a Config", color.LightBlue)
	if doConfig {
		if blueprint.BlueprintID == "" {
			bid, err := readString("What is the BlueprintID?", color.LightBlue, -1, nil)
			blueprint.BlueprintID = *bid
			if err != nil {
				log.Panic("Need the BlueprintID to create a config!")
			}
		}

		fileLoad := simpleQuestion("Load a User Config form a file?", color.LightBlue)
		if fileLoad {
			path, err := readString("Enter the path to the file you want to load:", color.LightBlue, -1, func(path string) bool {
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
				role, _ := readString("Enter a role name used of your VDC", color.LightBlue, -1, nil)
				roles = append(roles, *role)
				ctn := simpleQuestion("Add another role?", color.LightBlue)
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
				name, _ := readString("Enter a username ", color.LightBlue, -1, nil)
				pwd, _ := readString("Enter a password ", color.LightBlue, -1, nil)

				user.Username = *name
				user.Password = *pwd
				user.Roles = make([]string, 0)
				for {
					i := menu(fmt.Sprintf("Select role for %s",*name),color.LightBlue,user.Roles)
					user.Roles = append(user.Roles, roles[i])

					ctn := simpleQuestion("Add another role?", color.LightBlue)
					if !ctn {
						break
					}
				}
				users = append(users, user)

				ctn := simpleQuestion("Add another user?", color.LightBlue)
				if !ctn {
					break
				}
			}
			config.Users = users

		}
		apply := simpleQuestion(fmt.Sprintf("Do you want to commit this config?\n%+v", config), color.LightMagenta)
		if apply {
			err := client.sendConfig(config)
			if err != nil {
				log.Panic("Need the BlueprintID to create a config!")
			}
		}
	}
}

func blueprintCommand(blueprint BluePrint, client *ConfigClient) {
	createBlueprint := simpleQuestion("Do you want to initialized a new Blueprint?", color.LightBlue)
	if createBlueprint {
		fileLoad := simpleQuestion("Load a Blueprint Config form a File?", color.LightBlue)

		if fileLoad {
			path, err := readString("Enter the path to the file you want to load:", color.LightBlue, -1, func(path string) bool {
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
			bid, _ := readString("What is the BlueprintID?", color.LightBlue, -1, nil)
			blueprint.BlueprintID = *bid

			cid, _ := readString("What is the clientID?", color.LightBlue, -1, nil)
			blueprint.ClientId = *cid
		}

		apply := simpleQuestion(fmt.Sprintf("Do you want to commit this config?\n%+v", blueprint), color.LightMagenta)
		if apply {
			//TODO: print output!
			err := client.sendBlueprint(blueprint)
			if err != nil {
				log.Debug("failed to send blueprint")
				//TODO:
			}
		}

	}
}

