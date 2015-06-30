package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
)

func TestMain(m *testing.M) {
	os.MkdirAll("tmp", os.ModeDir|os.ModePerm)
	err := exec.Command("go", "build", "-o", "tmp/sqlfmt").Run()
	if err != nil {
		fmt.Println("Failed to build sqlfmt binary:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func sqlfmt(t *testing.T, sql string, args ...string) string {
	cmd := exec.Command("tmp/sqlfmt", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("cmd.StdinPipe failed: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("cmd.StdoutPipe failed: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		t.Fatalf("cmd.Start failed: %v", err)
	}

	_, err = fmt.Fprint(stdin, sql)
	if err != nil {
		t.Fatalf("fmt.Fprint failed: %v", err)
	}

	err = stdin.Close()
	if err != nil {
		t.Fatalf("stdin.Close failed: %v", err)
	}

	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(stdout) failed: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		t.Fatalf("cmd.Wait failed: %v", err)
	}

	return string(output)
}

func TestSqlFmt(t *testing.T) {
	tests := []struct {
		inputFile          string
		expectedOutputFile string
	}{
		{
			inputFile:          "simple_select_without_from.sql",
			expectedOutputFile: "simple_select_without_from.fmt.sql",
		},
		{
			inputFile:          "simple_select_with_from.sql",
			expectedOutputFile: "simple_select_with_from.fmt.sql",
		},
	}

	for i, tt := range tests {
		input, err := ioutil.ReadFile(path.Join("testdata", tt.inputFile))
		if err != nil {
			t.Fatal(err)
		}

		expected, err := ioutil.ReadFile(path.Join("testdata", tt.expectedOutputFile))
		if err != nil {
			t.Fatal(err)
		}

		output := sqlfmt(t, string(input))

		if output != string(expected) {
			actualFileName := path.Join("tmp", fmt.Sprintf("%d.sql", i))
			err = ioutil.WriteFile(actualFileName, []byte(output), os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}

			t.Errorf("%d. Given %s, did not receive %s. Unexpected output written to %s", i, tt.inputFile, tt.expectedOutputFile, actualFileName)
		}
	}
}