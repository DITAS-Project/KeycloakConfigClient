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
	"bufio"
	"fmt"
	"gopkg.in/gookit/color.v1"
	"os"
	"strconv"
)

var scanner = bufio.NewScanner(os.Stdin)

func simpleQuestion(question string, style color_v1.Color) bool {
	style.Println(question + "[y/n]")
	scanner.Scan()

	value := scanner.Text()
	if value == "y" || value == "Y" || value == "yes" || value == "YES" {
		return true
	}
	return false
}

func readString(question string, style color.Color, repeats int, validate func(string) bool) (*string, error) {
	var value string
	asked := -1
	for {
		if repeats < asked {
			break
		}
		asked++

		style.Println(question)
		scanner.Scan()
		value = scanner.Text()
		if validate != nil && validate(value) {
			return &value, nil
		} else {
			return &value, nil
		}
	}
	return nil, fmt.Errorf("no awnser recived")
}

func menu(question string, style color.Color, options []string) int {
	selector := ""
	for i, s := range options {
		selector += fmt.Sprintf("\t%d: %s\n", i, s);
	}

	for {
		num, err := readString(fmt.Sprintf("%s:\n %s", question, selector), style, 1, func(input string) bool {
			num, err := strconv.Atoi(input)
			return err == nil && num >= 0 && num < len(options)
		})
		if err == nil {
			i, _ := strconv.Atoi(*num)
			return i
		} else {
			color.LightRed.Println("Wrong Selection, please try again.")
		}
	}
}

