// osop
// Copyright (C) 2014 Karol 'Kenji Takahashi' Woźniak
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

func contains(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

func del(workspaces []*workspace, name string) []*workspace {
	for i, w := range workspaces {
		if w.Name == name {
			copy(workspaces[i:], workspaces[i+1:])
			workspaces[len(workspaces)-1] = nil
			return workspaces[:len(workspaces)-1]
		}
	}
	return workspaces
}

type event struct {
	EventName string
	Name      string
	Id        int
}

type workspace struct {
	Name       string
	Active     bool
	ActiveOn   int64
	Alerted    bool
	Layout     string
	Clients    uint
	HasClients bool
}

type Wingo struct {
	Workspaces []*workspace
	ActiveName string

	workspaces  map[string]*workspace
	clients     map[int]*workspace // FIXME: This name is ambiguous
	connection  net.Conn
	reader      *bufio.Reader
	eventReader *bufio.Reader
}

func (w *Wingo) GetEvented() interface{} {
	eventBytes, err := w.eventReader.ReadBytes(0)
	if err != nil {
		log.Printf("Error getting Wingo event: `%s`", err)
		return *w
	}
	var event event
	json.Unmarshal(eventBytes[:len(eventBytes)-1], &event)
	switch event.EventName {
	case "ChangedWorkspace", "ChangedVisibleWorkspace", "ChangedWorkspaceNames":
		w.updateWorkspaces()
	case "AddedWorkspace":
		w.Workspaces = append(w.Workspaces, w.getWorkspace(event.Name))
	case "RemovedWorkspace":
		w.Workspaces = del(w.Workspaces, event.Name)
	case "ChangedActiveClient":
		// FIXME: Active name should be per head, shouldn't it?
		w.ActiveName = w.getClientName(event.Id)
	case "MappedClient", "ManagedClient":
		w.addClient(event.Id, nil)
	case "UnmappedClient", "UnmanagedClient":
		w.removeClient(event.Id)
	}

	return *w
}

func (w *Wingo) getWorkspace(name string) *workspace {
	workspace := &workspace{Name: name}
	w.connection.Write([]byte(fmt.Sprintf("WorkspaceHead \"%s\"\x00", name)))
	head, err := w.reader.ReadString(0)
	if err != nil {
		log.Printf("Error getting Wingo workspace head: `%s`, `%s`", name, err)
	} else {
		workspace.ActiveOn, err = strconv.ParseInt(head[:len(head)-1], 0, 0)
		if err != nil {
			log.Printf("Error getting Wingo workspace head: `%s`, `%s`", name, err)
		} else {
			workspace.Active = workspace.ActiveOn != -1
		}
	}

	// TODO: This doesn't seem usable in current form
	w.connection.Write([]byte(fmt.Sprintf("GetLayout \"%s\"\x00", name)))
	layout, err := w.reader.ReadString(0)
	if err != nil {
		log.Printf("Error getting Wingo workspace layout: `%s`, `%s`", name, err)
	} else {
		workspace.Layout = layout[:len(layout)-1]
	}

	w.connection.Write([]byte(fmt.Sprintf("GetClientList \"%s\"\x00", name)))
	clients, err := w.reader.ReadString(0)
	if err != nil {
		log.Printf("Error getting Wingo workspace client list: `%s`, `%s`", name, err)
		return workspace
	}
	for _, client := range strings.Split(clients[:len(clients)-1], "\n") {
		if client == "" {
			continue
		}
		clientId, err := strconv.Atoi(client)
		if err != nil {
			log.Printf("Wrong Wingo client Id: `%s`, `%s`, `%s`", name, client, err)
			continue
		}
		w.addClient(clientId, workspace)
	}

	return workspace
}

func (w *Wingo) getClientName(id int) string {
	w.connection.Write([]byte(fmt.Sprintf("GetClientName %d\x00", id)))
	activeName, err := w.reader.ReadString(0)
	if err != nil {
		log.Printf("Error getting Wingo client name: `%d`, `%s`", id, err)
		return ""
	}
	return activeName[:len(activeName)-1]
}

func (w *Wingo) addClient(id int, workspace *workspace) {
	if workspace == nil {
		w.connection.Write([]byte(fmt.Sprintf("GetClientWorkspace %d\x00", id)))
		workspaceName, err := w.reader.ReadString(0)
		if err != nil {
			log.Printf("Error getting Wingo client workspace: `%d`, `%s`", id, err)
			return
		}
		workspace = w.workspaces[workspaceName[:len(workspaceName)-1]]
	}
	if workspace == nil {
		return
	}
	if w.clients[id] == workspace {
		return
	}
	workspace.Clients += 1
	workspace.HasClients = true
	w.clients[id] = workspace
}

func (w *Wingo) removeClient(id int) {
	workspace := w.clients[id]
	if workspace == nil {
		return
	}
	if workspace.Clients > 0 {
		workspace.Clients -= 1
		workspace.HasClients = workspace.Clients != 0
	}
	delete(w.clients, id)
}

func (w *Wingo) updateWorkspaces() {
	w.connection.Write([]byte("GetWorkspaceList\x00"))
	workspacesString, err := w.reader.ReadString(0)
	if err != nil {
		log.Printf("Error getting Wingo workspace list: `%s`", err)
		return
	}

	workspaces := strings.Split(workspacesString[:len(workspacesString)-1], "\n")

	w.Workspaces = make([]*workspace, len(workspaces))
	w.workspaces = make(map[string]*workspace, len(workspaces))
	for i, workspaceName := range workspaces {
		wrkspace := w.getWorkspace(workspaceName)
		w.Workspaces[i] = wrkspace
		w.workspaces[workspaceName] = wrkspace
	}
}

func NewWingo(config config) interface{} {
	wingoCmd := exec.Command("wingo", "--show-socket")
	socketBytes, err := wingoCmd.Output()
	for err != nil {
		log.Println("Problem getting wingo socket location, retrying:", err)
		wingoCmd := exec.Command("wingo", "--show-socket") // TODO: Refactor
		socketBytes, err = wingoCmd.Output()
	}
	socket := string(socketBytes)[:len(socketBytes)-1]
	conn, err := net.Dial("unix", socket)
	for err != nil {
		log.Println("Problem connecting to wingo socket, retrying:", err)
		conn, err = net.Dial("unix", socket) // TODO: Refactor
	}
	evconn, err := net.Dial("unix", socket+"-notify")
	for err != nil {
		log.Println("Problem connecting to wingo-notify socket, retrying:", err)
		evconn, err = net.Dial("unix", socket+"-notify") // TODO: Same
	}

	wingo := &Wingo{
		connection:  conn,
		reader:      bufio.NewReader(conn),
		eventReader: bufio.NewReader(evconn),
	}
	wingo.clients = map[int]*workspace{}
	wingo.updateWorkspaces()
	return wingo
}

func init() {
	registry.AddReceiver("Wingo", NewWingo)
}
