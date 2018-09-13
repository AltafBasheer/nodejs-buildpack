package v3

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	libbuildpackV3 "github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var _ = Describe("Nodejs V3 buildpack", func() {
	It("should run V3 detection", func() {
		bpDir, err := cutlass.FindRoot()
		Expect(err).ToNot(HaveOccurred())

		workingDir, err := ioutil.TempDir("/tmp", "")
		Expect(err).ToNot(HaveOccurred())
		defer os.RemoveAll(workingDir)

		appDir := filepath.Join(workingDir, "app")
		err = os.Mkdir(appDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		err = libbuildpack.CopyDirectory(filepath.Join(bpDir, "fixtures", "simple_app"), appDir)
		Expect(err).ToNot(HaveOccurred())

		output := &bytes.Buffer{}
		cmd := exec.Command(
			"docker",
			"run",
			"--rm",
			"-v",
			fmt.Sprintf("%s:/workspace", workingDir),
			"-v",
			fmt.Sprintf("%s:/buildpacks/%s/latest", bpDir, "org.cloudfoundry.buildpacks.nodejs"),
			"bpv3:build",
			"/lifecycle/detector",
			"-order",
			"/buildpacks/org.cloudfoundry.buildpacks.nodejs/latest/fixtures/v3/order.toml",
			"-group",
			"/workspace/group.toml",
			"-plan",
			"/workspace/plan.toml",
		)
		cmd.Stdout = output
		cmd.Stderr = output
		if err = cmd.Run(); err != nil {
			Fail("failed to run V3 detection " + output.String())
		}

		group := struct {
			Buildpacks []struct {
				Id      string `toml:"id"`
				Version string `toml:"version"`
			} `toml:"buildpacks"`
		}{}
		_, err = toml.DecodeFile(filepath.Join(workingDir, "group.toml"), &group)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(group.Buildpacks)).To(Equal(1))
		Expect(group.Buildpacks[0].Id).To(Equal("org.cloudfoundry.buildpacks.nodejs"))
		Expect(group.Buildpacks[0].Version).To(Equal("1.6.32"))

		plan := libbuildpackV3.BuildPlan{}
		_, err = toml.DecodeFile(filepath.Join(workingDir, "plan.toml"), &plan)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(plan)).To(Equal(1))
		Expect(plan).To(HaveKey("node"))
		Expect(plan["node"].Version).To(Equal("10.10.0"))
	})
})
