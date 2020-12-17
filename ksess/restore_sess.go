/* restore_sess

 */
package ksess

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	//"regexp"

	"mak_common/kerr"
	//"kot_common/kutils"

	"github.com/gorilla/securecookie"
)

const (
	hashKeyLen  = 64
	blockKeylen = 32
)

//restoreSessions is the main procedure that restores saved sessions if it is possible
//or(!) generates and saved new session keys that provides possibility for future restoring sessions.
//The error may be in both cases and indicates no successful performing.
//The possibility is determed by restorePossible function
func restoreSessions() (err error) {
	kerr.PrintDebugMsg(false, "restoreSess", "ksess. restoreSessions HERE")
	var ok bool
	var sD sharedData

	ok, err = sessDirInInitPos()
	if err != nil {
		err = fmt.Errorf("ksess.restoreSessions: err=%v", err.Error())
		return
	}
	if ok {
		newHashKey := securecookie.GenerateRandomKey(hashKeyLen)
		newBlockKey := securecookie.GenerateRandomKey(blockKeylen)
		cookieHandler = securecookie.New(
			newHashKey,
			newBlockKey,
		)
		if err = saveSessKeys(newHashKey, newBlockKey); err != nil {
			err = fmt.Errorf("ksess.restoreSessions: saveSessKeys(newHashKey, newBlockKey) err=%v", err.Error())
			return
		}

		sD.SessionKeysGeneratedAt = time.Now().Format(startFormat)
		if err = saveSharedData(sD); err != nil {
			err = fmt.Errorf("ksess.restoreSessions: saveSharedData (initial saving) err=%v", err.Error())
			return
		}

		return //!190704
	}

	if err = restoreSessionsKeys(); err != nil {
		err = fmt.Errorf("ksess.restoreSessions: restoreSessionsKeys err=%v", err.Error())
		return
	}

	if err = restoreSessionUsers(); err != nil {
		err = fmt.Errorf("ksess.restoreSessions: restoreSessionUsers err=%v", err.Error())
		return
	}

	if err = updateSharedData(); err != nil {
		err = fmt.Errorf("ksess.restoreSessions: updating shared data err=%v", err.Error())
		return
	}

	return
}

//sessDirInInitPos stands for the sessions directory in initial position
//That is session keys have not been save still and session/session (saved sessions list) is empty.
//Not nil err indicates fatal error - impossibility normal working. In that case the ok is undefined
//sessDirInInitPos returns not nil err if there are not session keys but the list is not empty.
func sessDirInInitPos() (ok bool, err error) {
	var fileName string
	var f *os.File
	var fI os.FileInfo
	var listDirContent []os.FileInfo
	var isBlockKey, isHashKey, isListDirEmpty bool

	//defer func() {
	//	SendToGenLog("ksess.sessDirInInitPos", fmt.Sprintf("ok=%v; err=%v", ok, err))
	//}()

	fileName = "sessions"
	if fI, err = os.Stat(fileName); err != nil {
		err = fmt.Errorf("ksess.sessDirInInitPos: There is not the keys direstory (sessions)")
		goto writeToLog // return
	} else {
		if !fI.IsDir() {
			err = fmt.Errorf("ksess.sessDirInInitPos: the sessions must be a directory (the keys directory)")
			goto writeToLog //return
		}
	}

	fileName = fmt.Sprintf("sessions%vsessions", string(os.PathSeparator))
	if fI, err = os.Stat(fileName); err != nil {
		err = fmt.Errorf("ksess.sessDirInInitPos: There is not the list direstory (sessions/sessions)")
		goto writeToLog //return
	} else {
		if !fI.IsDir() {
			err = fmt.Errorf("ksess.sessDirInInitPos: the sessions/sessions must be a directory (the list directory)")
			goto writeToLog //return
		}
	}

	fileName = fmt.Sprintf("sessions%vblockKey", string(os.PathSeparator))
	if fI, err = os.Stat(fileName); err != nil {
		isBlockKey = false
	} else {
		if !fI.IsDir() && fI.Size() == blockKeylen {
			isBlockKey = true
		} else {
			err = fmt.Errorf("ksess.sessDirInInitPos: the sessions/blockKey has not proper size or is a directory")
			goto writeToLog //return
		}
	}

	fileName = fmt.Sprintf("sessions%vhashKey", string(os.PathSeparator))
	if fI, err = os.Stat(fileName); err != nil {
		isHashKey = false
	} else {
		if !fI.IsDir() && fI.Size() == hashKeyLen {
			isHashKey = true
		} else {
			err = fmt.Errorf("ksess.sessDirInInitPos: the sessions/hashKey has not proper size or is a directory")
			goto writeToLog //return
		}
	}

	fileName = fmt.Sprintf("sessions%vsessions", string(os.PathSeparator))
	if f, err = os.Open(fileName); err != nil {
		err = fmt.Errorf("ksess.sessDirInInitPos: open list direstory err=%v", err.Error())
		return
	} else {
		if listDirContent, err = f.Readdir(0); err != nil {
			err = fmt.Errorf("ksess.sessDirInInitPos: reading list direstory err=%v", err.Error())
			goto writeToLog //return
		} else {
			if len(listDirContent) == 0 {
				isListDirEmpty = true
			} else {
				isListDirEmpty = false
			}
		}
	}

	kerr.PrintDebugMsg(false, "restoreSess", fmt.Sprintf("ksess. sessDirInInitPos isBlockKey=%v; isHashKey=%v", isBlockKey, isHashKey))

	if (isBlockKey && !isHashKey) || (!isBlockKey && isHashKey) {
		err = fmt.Errorf("ksess.sessDirInInitPos: the key files must be exist or not exist simultaneously")
		goto writeToLog //return
	}
	if !isBlockKey && !isHashKey && isListDirEmpty {
		ok = true
	} else {
		ok = false
	}

writeToLog:
	SendToGenLog("ksess.sessDirInInitPos", fmt.Sprintf("ok=%v; err=%v", ok, err))
	return
}

func saveSess(client *sessClient) {
	var marshaledClient []byte
	var marshaledClientByff *bytes.Buffer
	var err error
	var f *os.File
	var fName string
	if marshaledClient, err = json.Marshal(*client); err != nil {
		kerr.SysErrPrintf("ksess.saveSess:user=%v(%v); err=%v", client.User_ID, client.Tag, err.Error())
		return
	}
	fName = fmt.Sprintf("sessions%vsessions%v%v.json", string(os.PathSeparator), string(os.PathSeparator), client.User_ID)
	kerr.PrintDebugMsg(false, "restoreSess", fmt.Sprintf("ksess.saveSess:fName=%v", fName))
	if f, err = os.Create(fName); err != nil {
		kerr.SysErrPrintf("ksess.saveSess: file create err=%v", err.Error())
		return
	}
	defer f.Close()

	marshaledClientByff = bytes.NewBuffer(marshaledClient)
	if _, err = io.Copy(f, marshaledClientByff); err != nil {
		kerr.SysErrPrintf("ksess.saveSess: io.Copy(f, marshaledClientByff) err=%v", err.Error())
		return
	}
	kerr.PrintDebugMsg(false, "restoreSess", "ksess.saveSess: returning without errors")
}

func deleteSavedSess(userId int) {
	var fName = fmt.Sprintf("sessions%vsessions%v%v.json", string(os.PathSeparator), string(os.PathSeparator), userId)
	var err error
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		return
	}
	if err = os.Remove(fName); err != nil {
		kerr.SysErrPrintf("ksess.deleteSavedSess: (%v) err=%v", fName, err.Error())
		return
	}
}

func saveSessKeys(hashKey, blockKey []byte) (err error) {
	var hashKeyFName = fmt.Sprintf("sessions%vhashKey", string(os.PathSeparator))
	var blockKeyFName = fmt.Sprintf("sessions%vblockKey", string(os.PathSeparator))
	var hashKeyF, blockKeyF *os.File
	var hashKeyBuff, blockKeyBuff *bytes.Buffer
	if hashKeyF, err = os.Create(hashKeyFName); err != nil {
		err = fmt.Errorf("ksess.saveSessKeys: hashKeyF create err=%v", err.Error())
		return
	}
	defer hashKeyF.Close()
	hashKeyBuff = bytes.NewBuffer(hashKey)
	if _, err = io.Copy(hashKeyF, hashKeyBuff); err != nil {
		err = fmt.Errorf("ksess.saveSessKeys: io.Copy(hashKeyF, hashKeyBuff) err=%v", err.Error())
		return
	}

	if blockKeyF, err = os.Create(blockKeyFName); err != nil {
		err = fmt.Errorf("ksess.saveSessKeys: blockKeyF create err=%v", err.Error())
		return
	}
	defer blockKeyF.Close()
	blockKeyBuff = bytes.NewBuffer(blockKey)
	if _, err = io.Copy(blockKeyF, blockKeyBuff); err != nil {
		err = fmt.Errorf("ksess.saveSessKeys: io.Copy(blockKeyF, blockKeyBuff) err=%v", err.Error())
		return
	}

	SendToGenLog("saveSessKeys", "saved new keys")
	return
}

func restoreSessionsKeys() (err error) {
	var hashKeyFName = fmt.Sprintf("sessions%vhashKey", string(os.PathSeparator))
	var blockKeyFName = fmt.Sprintf("sessions%vblockKey", string(os.PathSeparator))
	var hashKeyF, blockKeyF *os.File
	var fSize int64
	var hashKeyBuff, blockKeyBuff bytes.Buffer

	if hashKeyF, err = os.Open(hashKeyFName); err != nil {
		err = fmt.Errorf("restoreSessionsKeys: os.Open(hashKeyFName) err=%v", err.Error())
		return
	}
	defer hashKeyF.Close()
	if blockKeyF, err = os.Open(blockKeyFName); err != nil {
		err = fmt.Errorf("restoreSessionsKeys: os.Open(blockKeyFName) err=%v", err.Error())
		return
	}
	defer blockKeyF.Close()
	if fSize, err = io.Copy(&blockKeyBuff, blockKeyF); err != nil {
		err = fmt.Errorf("restoreSessionsKeys: io.Copy(blockKeyBuff, blockKeyF) err=%v", err.Error())
		return
	}
	if fSize != blockKeylen {
		err = fmt.Errorf("restoreSessionsKeys: blockKeyLen!=%v", blockKeylen)
		return
	}
	if fSize, err = io.Copy(&hashKeyBuff, hashKeyF); err != nil {
		err = fmt.Errorf("restoreSessionsKeys: io.Copy(hashKeyBuff, hashKeyF) err=%v", err.Error())
		return
	}
	if fSize != hashKeyLen {
		err = fmt.Errorf("restoreSessionsKeys: hashKeyLen!=%v", hashKeyLen)
		return
	}

	cookieHandler = securecookie.New(
		hashKeyBuff.Bytes(),
		blockKeyBuff.Bytes(),
	)

	SendToGenLog("restoreSessionsKeys", "restored")
	return

}

func restoreSessionUsers() (err error) {
	var sessDirName string
	var sessDirF *os.File
	var sessNames []string

	sessDirName = fmt.Sprintf("sessions%vsessions", string(os.PathSeparator))
	if sessDirF, err = os.Open(sessDirName); err != nil {
		err = fmt.Errorf("restoreSessionUsers:os.Open(sessDirName) err=%v", err.Error())
		return
	}
	defer sessDirF.Close()

	if sessNames, err = sessDirF.Readdirnames(0); err != nil {
		err = fmt.Errorf("restoreSessionUsers: Readdirnames(0) err=%v", err.Error())
		return
	}
	for i := 0; i < len(sessNames); i++ {
		var userFileName string
		var cln sessClient
		var clnF *os.File
		var clnBuff bytes.Buffer
		userFileName = fmt.Sprintf("sessions%vsessions%v%v", string(os.PathSeparator), string(os.PathSeparator), sessNames[i])
		fmt.Printf("--M-- userFileName=%v\n", userFileName)
		if clnF, err = os.Open(userFileName); err != nil {
			kerr.SysErrPrintf("restoreSessionUsers: os.Open(%v) err=%v", userFileName, err.Error())
			continue
		}
		if _, err = io.Copy(&clnBuff, clnF); err != nil {
			kerr.SysErrPrintf("restoreSessionUsers: io.Copy err=%v", err.Error())
			continue
		}
		if err = json.Unmarshal(clnBuff.Bytes(), &cln); err != nil {
			kerr.SysErrPrintf("restoreSessionUsers:json.Unmarshal(%v) err=%v", userFileName, err.Error())
			continue
		} else {
			cln.hub = hub
			cln.send = make(chan []byte)
			hub.registerSess(&cln)
			//SendToGenLog("restoreSessionUsers", cln.String("; "))
		}

	}
	return
}

func updateSavedSess(client *sessClient) {
	var marshaledClient []byte
	var marshaledClientBuff *bytes.Buffer
	var err error
	var f *os.File
	var fName string

	fName = fmt.Sprintf("sessions%vsessions%v%v.json", string(os.PathSeparator), string(os.PathSeparator), client.User_ID)
	if f, err = os.Create(fName); err != nil {
		kerr.SysErrPrintf("ksess.updateSavedSess: file open err=%v", err.Error())
		return
	}
	defer f.Close()

	if marshaledClient, err = json.Marshal(*client); err != nil {
		kerr.SysErrPrintf("ksess.updateSavedSess:user=%v(%v); err=%v", client.User_ID, client.Tag, err.Error())
		return
	}

	marshaledClientBuff = bytes.NewBuffer(marshaledClient)
	if _, err = io.Copy(f, marshaledClientBuff); err != nil {
		kerr.SysErrPrintf("ksess.updateSavedSess: io.Copy(f, marshaledClientByff) err=%v", err.Error())
		return
	}
	//kerr.PrintDebugMsg(false, "restoreSess", "ksess.saveSess: returning without errors")
}

//190703_04--------------------------------
//sharedData stands for data shared by all

type sharedData struct {
	SessionKeysGeneratedAt  string
	LastRestoringOccurredAt string
}

func saveSharedData(sd sharedData) (err error) {
	var fName = fmt.Sprintf("sessions%vsharedData.json", string(os.PathSeparator))
	var f *os.File
	var buff *bytes.Buffer
	var marshaledData []byte
	if f, err = os.Create(fName); err != nil {
		err = fmt.Errorf("ksess.saveSharedData: os.Create(fName) err=%v", err.Error())
		return
	}
	defer f.Close()
	if marshaledData, err = json.Marshal(sd); err != nil {
		err = fmt.Errorf("ksess.saveSharedData: json.Marshal err=%v", err.Error())
		return
	}
	buff = bytes.NewBuffer(marshaledData)
	if _, err = io.Copy(f, buff); err != nil {
		err = fmt.Errorf("ksess.saveSharedData: io.Copy err=%v", err.Error())
		return
	}
	return
}

func updateSharedData() (err error) {
	var fName = fmt.Sprintf("sessions%vsharedData.json", string(os.PathSeparator))
	var fOld *os.File
	var buff *bytes.Buffer
	var savedSD sharedData
	//var newSDSlice []byte

	if fOld, err = os.Open(fName); err != nil {
		err = fmt.Errorf("ksess.updateSharedData:os.Open(fOld) err=%v", err.Error())
		return
	}
	buff = bytes.NewBuffer(nil)
	if _, err = io.Copy(buff, fOld); err != nil {
		err = fmt.Errorf("ksess.saveSharedData: io.Copy err=%v", err.Error())
		fOld.Close()
		return
	}
	fOld.Close()
	//kerr.PrintDebugMsg(false, "restoreSess", fmt.Sprintf("buff.Bytes=%v", string(buff.Bytes())))

	if err = json.Unmarshal(buff.Bytes(), &savedSD); err != nil {
		err = fmt.Errorf("ksess.updateSharedData:json.Unmarshal savedSD err=%v", err.Error())
		return
	}

	savedSD.LastRestoringOccurredAt = time.Now().Format(startFormat)
	if err = saveSharedData(savedSD); err != nil {
		err = fmt.Errorf("ksess.updateSharedData: saving updated err=%v", err.Error())
		return
	}

	return
}
