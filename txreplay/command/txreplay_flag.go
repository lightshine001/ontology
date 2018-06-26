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

package command

import (
	"strings"

	"github.com/urfave/cli"
)

const (
	DEFAULT_TX_EXPORT_FILE = "./txs.dat"
)

var (
	// Tx export and import
	TxExportFileFlag = cli.StringFlag{
		Name:  "file",
		Usage: "Path of export file",
		Value: DEFAULT_TX_EXPORT_FILE,
	}

	TxExportHeightFlag = cli.UintFlag{
		Name:  "height",
		Usage: "Using to specifies the beginning of the block to be exported.",
		Value: 0,
	}

	ImportTxFileFlag = cli.StringFlag{
		Name:  "importtxsfile",
		Usage: "Path of import txs file",
		Value: DEFAULT_TX_EXPORT_FILE,
	}

	HostIPFlag = cli.StringFlag{
		Name:  "ip",
		Usage: "node's ip address",
		Value: "localhost",
	}

	RPCPortFlag = cli.UintFlag{
		Name:  "rpcport",
		Usage: "Json rpc server listening port",
		Value: 20336,
	}
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
