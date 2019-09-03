package template

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	daemon2 "proctor/internal/app/cli/daemon"
	"proctor/internal/app/cli/utility/io"
	procMetadata "proctor/internal/pkg/model/metadata"
	"proctor/internal/pkg/model/metadata/env"
)

type TemplateCmdTestSuite struct {
	suite.Suite
	mockPrinter        *io.MockPrinter
	mockProctorDClient *daemon2.MockClient
	testTemplateCmd    *cobra.Command
}

func (s *TemplateCmdTestSuite) SetupTest() {
	s.mockPrinter = &io.MockPrinter{}
	s.mockProctorDClient = &daemon2.MockClient{}
	s.testTemplateCmd = NewCmd(s.mockPrinter, s.mockProctorDClient)
}

func (s *TemplateCmdTestSuite) TestTemplateCmdUsage() {
	assert.Equal(s.T(), "template", s.testTemplateCmd.Use)
}

func (s *TemplateCmdTestSuite) TestTemplateCmdHelp() {
	assert.Equal(s.T(), "Get input template of a procs", s.testTemplateCmd.Short)
	assert.Equal(s.T(), "To get input template of a procs, this command retrieve an example template derived from stored metadata", s.testTemplateCmd.Long)
	assert.Equal(s.T(), "proctor template say-hello-world say-hello-world.yaml", s.testTemplateCmd.Example)
}

func (s *TemplateCmdTestSuite) TestTemplateCmdRun() {
	t := s.T()

	filename := "/tmp/yaml-test-template"
	defer os.Remove(filename)

	arg := env.VarMetadata{
		Name:        "arg-one",
		Description: "arg one description",
	}

	secret := env.VarMetadata{
		Name:        "secret-one",
		Description: "secret one description",
	}

	anyProc := procMetadata.Metadata{
		Name:             "do-something",
		Description:      "does something",
		Contributors:     "user@example.com",
		Organization:     "org",
		AuthorizedGroups: []string{"group_one", "group_two"},
		EnvVars: env.Vars{
			Args:    []env.VarMetadata{arg},
			Secrets: []env.VarMetadata{secret},
		},
	}
	procList := []procMetadata.Metadata{anyProc}

	s.mockProctorDClient.On("ListProcs").Return(procList, nil).Once()

	s.mockPrinter.On("Println", "\nArgs", color.FgMagenta).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("%-40s %-100s", arg.Name, arg.Description), color.Reset).Once()
	s.mockPrinter.On("Println", fmt.Sprintf("\nTo %s, run:\nproctor execute %s -f %s ARG_ONE=foo ARG_TWO=bar", anyProc.Name, anyProc.Name, filename), color.FgGreen).Once()

	s.testTemplateCmd.Run(&cobra.Command{}, []string{anyProc.Name, filename})

	templateFile, err := os.Open(filename)
	assert.NoError(t, err)
	defer templateFile.Close()

	templateBuffer, err := ioutil.ReadAll(templateFile)
	assert.Equal(t, templateBuffer, []byte("# arg one description\narg-one:\n"))

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *TemplateCmdTestSuite) TestTemplateCmdForIncorrectUsage() {
	s.mockPrinter.On("Println", "Incorrect command. See `proctor template --help` for usage", color.FgRed).Once()

	s.testTemplateCmd.Run(&cobra.Command{}, []string{})

	s.mockPrinter.AssertExpectations(s.T())
}

func (s *TemplateCmdTestSuite) TestTemplateCmdRunProctorDClientFailure() {
	filename := "/tmp/yaml-test-template"

	s.mockProctorDClient.On("ListProcs").Return([]procMetadata.Metadata{}, errors.New("test error")).Once()
	s.mockPrinter.On("Println", "test error", color.FgRed).Once()

	s.testTemplateCmd.Run(&cobra.Command{}, []string{"do-something", filename})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func (s *TemplateCmdTestSuite) TestTemplateCmdRunProcNotSupported() {
	filename := "/tmp/yaml-test-template"

	s.mockProctorDClient.On("ListProcs").Return([]procMetadata.Metadata{}, nil).Once()
	testProcName := "do-something"
	s.mockPrinter.On("Println", fmt.Sprintf("Proctor doesn't support Proc `%s`\nRun `proctor list` to view supported Procs", testProcName), color.FgRed).Once()

	s.testTemplateCmd.Run(&cobra.Command{}, []string{testProcName, filename})

	s.mockProctorDClient.AssertExpectations(s.T())
	s.mockPrinter.AssertExpectations(s.T())
}

func TestTemplateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(TemplateCmdTestSuite))
}
