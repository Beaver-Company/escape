/*
Copyright 2017, 2018 Ankyra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"github.com/ankyra/escape/model"
	. "github.com/ankyra/escape/model/interfaces"
)

type PackageController struct{}

func (PackageController) Package(context Context, forceOverwrite bool) error {
	context.PushLogRelease(context.GetReleaseMetadata().GetQualifiedReleaseId())
	context.PushLogSection("Package")
	context.Log("package.start", nil)
	archiver := model.NewReleaseArchiver()
	releasePath, err := archiver.Archive(context.GetReleaseMetadata(), forceOverwrite)
	if err != nil {
		return err
	}
	context.Log("package.finished", map[string]string{
		"path": releasePath,
	})
	context.PopLogRelease()
	context.PopLogSection()
	return nil
}
