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

package req

import (
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	cactor "github.com/ontio/ontology/consensus/actor/msg"
)

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}

func NotifyEmergencyGovCmd(cmd interface{}) {
	if ConsensusPid != nil {
		ConsensusPid.Tell(cmd)
	}
}

func ConsensusSrvStart() error {
	future := ConsensusPid.RequestFuture(&cactor.StartConsensus{}, time.Second*10)
	_, err := future.Result()
	if err != nil {
		return err
	}
	return nil
}
func ConsensusSrvHalt() error {
	future := ConsensusPid.RequestFuture(&cactor.StopConsensus{}, time.Second*10)
	_, err := future.Result()
	if err != nil {
		return err
	}
	return nil
}
