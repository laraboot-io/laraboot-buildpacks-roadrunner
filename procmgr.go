/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package roadrunner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/paketo-buildpacks/php-web/procmgr"
)

type procMsg struct {
	ProcName string
	Cmd      *exec.Cmd
	Err      error
}

func RunProcs(procs procmgr.Procs) error {
	msgs := make(chan procMsg)

	for procName, proc := range procs.Processes {
		go runProc(procName, proc, msgs)
	}

	msg := <-msgs
	_, _ = fmt.Fprintln(os.Stderr, "process", msg.ProcName, "exited, status:", msg.Cmd.ProcessState)
	return msg.Err
}

func runProc(procName string, proc procmgr.Proc, msgs chan procMsg) {
	cmd := exec.Command(proc.Command, proc.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		msgs <- procMsg{procName, cmd, err}
	}

	err = cmd.Wait()
	if err != nil {
		msgs <- procMsg{procName, cmd, err}
	}

	msgs <- procMsg{procName, cmd, nil}
}
