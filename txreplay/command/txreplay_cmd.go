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
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/urfave/cli"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/txreplay/utils"
)

var TxExportCommand = cli.Command{
	Name:      "txexport",
	Usage:     "Export txs in DB to a file",
	ArgsUsage: "",
	Action:    exportTxs,
	Flags: []cli.Flag{
		HostIPFlag,
		RPCPortFlag,
		TxExportFileFlag,
		TxExportHeightFlag,
	},
	Description: "",
}

func exportTxs(ctx *cli.Context) error {
	ip := ctx.String(GetFlagName(HostIPFlag))
	port := ctx.Uint(GetFlagName(RPCPortFlag))
	utils.SetIPPort(ip, port)

	txFile := ctx.String(GetFlagName(TxExportFileFlag))
	if txFile == "" {
		fmt.Printf("Missing file argumen\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if common.FileExisted(txFile) {
		return fmt.Errorf("File:%s has already exist", txFile)
	}
	startHeight := ctx.Uint(GetFlagName(TxExportHeightFlag))
	blockCount, err := utils.GetBlockCount()
	if err != nil {
		return fmt.Errorf("GetBlockCount error:%s", err)
	}

	if startHeight >= uint(blockCount) {
		return fmt.Errorf("The specified height is over current height")
	}

	ef, err := os.OpenFile(txFile, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return fmt.Errorf("Open file:%s error:%s", txFile, err)
	}
	defer ef.Close()
	fWriter := bufio.NewWriter(ef)

	totalBlocks := int(blockCount) - int(startHeight)
	uiprogress.Start()
	bar := uiprogress.AddBar(totalBlocks).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Remaining Block %d", totalBlocks-b.Current())
		})

	fmt.Printf("Start export...\n")
	var count uint64
	for i := uint32(startHeight); i < uint32(blockCount); i++ {
		blockData, err := utils.GetBlockData(i)
		if err != nil {
			return fmt.Errorf("Get block:%d error:%s", i, err)
		}
		var block types.Block
		buf := bytes.NewBuffer(blockData)
		err = block.Deserialize(buf)
		if err != nil {
			return fmt.Errorf("failed to read block at height %d err %v", i, err)
		}
		fWriter.WriteString(fmt.Sprintf("Block %d num %d\n", i, len(block.Transactions)))
		for _, tx := range block.Transactions {
			txHex := hex.EncodeToString(tx.ToArray())
			_, err = fWriter.WriteString(fmt.Sprintf("%x %s\n", tx.Hash(), txHex))
			if err != nil {
				return fmt.Errorf("failed to write tx data %x at block height %d", tx.Hash(), i)
			}
			count++
		}
		bar.Incr()
	}
	uiprogress.Stop()

	err = fWriter.Flush()
	if err != nil {
		return fmt.Errorf("Export flush file error:%s", err)
	}
	fmt.Printf("Export txs successfully.\n")
	fmt.Printf("Total txs:%d from block %d to block %d\n", count, startHeight, blockCount)
	fmt.Printf("Export file:%s\n", txFile)
	return nil
}

var TxImportCommand = cli.Command{
	Name:      "tximport",
	Usage:     "Import txs from a file",
	ArgsUsage: "",
	Action:    importTxs,
	Flags: []cli.Flag{
		HostIPFlag,
		RPCPortFlag,
		ImportTxFileFlag,
	},
	Description: "",
}

func importTxs(ctx *cli.Context) {
	ip := ctx.String(GetFlagName(HostIPFlag))
	port := ctx.Uint(GetFlagName(RPCPortFlag))
	utils.SetIPPort(ip, port)

	txFile := ctx.String(GetFlagName(ImportTxFileFlag))
	if txFile == "" {
		fmt.Println("Missing file argument")
		cli.ShowSubcommandHelp(ctx)
		return
	}

	ifile, err := os.OpenFile(txFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ifile.Close()

	fReader := bufio.NewReader(ifile)

	fmt.Printf("%s Start import Txs...\n",
		time.Now().UTC().Format(time.UnixDate))

	var txs []string
	count := 0
	summary := 0
	errNum := 0
	rateLimiter := time.NewTicker(time.Millisecond * 1)
	for {
		line, err := fReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if strings.HasPrefix(line, "Block ") {
			for _, tx := range txs {
				err := utils.SendRawTransaction(tx)
				if err != nil {
					fmt.Printf("failed to send rpc reqeust. tx %s err %v\n", tx, err)
					errNum++
					continue
				}
				summary++
				<-rateLimiter.C
			}
			if len(txs) != 0 {
				fmt.Printf("%s Sent tx count %d, errNum %d\n",
					time.Now().UTC().Format(time.UnixDate), summary, errNum)
			}
			index := strings.LastIndex(line, " ")
			num, _ := strconv.Atoi(line[index+1 : len(line)-1])
			txs = make([]string, 0, num)
		} else {
			index := strings.Index(line, " ")
			if index < 0 {
				fmt.Printf("failed to split tx %s\n", line)
				continue
			}
			txHex := line[index+1 : len(line)-1]
			txs = append(txs, txHex)
			count++
		}
	}

	for _, tx := range txs {
		err := utils.SendRawTransaction(tx)
		if err != nil {
			fmt.Printf("failed to send rpc reqeust. tx %s err %v\n", tx, err)
			errNum++
			continue
		}
		summary++
		<-rateLimiter.C
	}

	fmt.Printf("%s Import Txs complete, total txs %d sent txs %d errNum %d\n",
		time.Now().UTC().Format(time.UnixDate), count, summary, errNum)
	rateLimiter.Stop()
}
