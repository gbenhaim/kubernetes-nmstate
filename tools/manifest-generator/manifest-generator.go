/*
 * Copyright 2019 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"os"
	"text/template"

	"github.com/golang/glog"
)

type templateData struct {
	Namespace         string
	ImageRegistry     string
	ImageTag          string
	PullPolicy        string
	StateHandlerImage string
}

var (
	namespace         = flag.String("namespace", "", "")
	imageRegistry     = flag.String("image-registry", "", "")
	imageTag          = flag.String("image-tag", "", "")
	pullPolicy        = flag.String("pull-policy", "", "")
	stateHandlerImage = flag.String("state-handler-image", "", "")
)

func main() {
	templFile := flag.String("template", "", "")
	flag.Parse()

	if *templFile == "" {
		panic("please provide path to a template file")
	}

	generateFromFile(*templFile)
}

func generateFromFile(templFile string) {
	data := &templateData{
		Namespace:         *namespace,
		ImageRegistry:     *imageRegistry,
		ImageTag:          *imageTag,
		PullPolicy:        *pullPolicy,
		StateHandlerImage: *stateHandlerImage,
	}

	file, err := os.OpenFile(templFile, os.O_RDONLY, 0)
	if err != nil {
		glog.Fatalf("Failed to open file %s: %v\n", templFile, err)
	}
	defer file.Close()

	tmpl := template.Must(template.ParseFiles(templFile))
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		glog.Fatalf("Error executing template: %v\n", err)
	}
}
