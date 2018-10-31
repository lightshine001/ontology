/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package netserver

import (
	"fmt"
	"testing"
	//"time"

	"github.com/ontio/ontology/common/log"
)

func init() {
	log.Init(log.Stdout)
	fmt.Println("Start test the netserver...")
}

func TestNewNetworkManager(t *testing.T) {
	networkManager1 := NewNetworkManager(20338)
	networkManager2 := NewNetworkManager(20339)
	//log.Info(networkManager)
	//networkManager1.base.SetPort(20338)
	//networkManager2.base.SetPort(20339)
	log.Infof("id :%v", networkManager1.base.GetID())
	log.Infof("version: %v", networkManager1.base.GetPort())
	dest := fmt.Sprintf("/ip4/127.0.0.1/tcp/20338/ipfs/%s", networkManager1.base.GetID().Pretty())
	networkManager2.Connect(dest)
}
