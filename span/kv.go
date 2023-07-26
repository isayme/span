package span

import (
	"bytes"
	"io/fs"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

var dbFileMode fs.FileMode = 0600

var dbFilePath = "span.db"

const bucketNameSpan = "span"
const bucketKeySalt = "salt"
const bucketKeyEncryptMasterKey = "encryptMasterKey"
const bucketKeyAuthKey = "authKey"

func openBolt() (*bolt.DB, error) {
	db, err := bolt.Open(dbFilePath, dbFileMode, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "打开DB文件失败: %s", dbFilePath)
	}
	return db, nil
}

func InitBolt(path string) error {
	if path != "" {
		dbFilePath = path
	}

	db, err := openBolt()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketNameSpan))
		if err != nil {
			return errors.Wrap(err, "DB初始化失败")
		}
		return nil
	})
	return err
}

func ReadBolt() (salt, encryptMasterKey, authKey []byte, err error) {
	db, err := openBolt()
	if err != nil {
		return nil, nil, nil, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketNameSpan))

		salt = bytes.Clone(b.Get([]byte(bucketKeySalt)))
		encryptMasterKey = bytes.Clone(b.Get([]byte(bucketKeyEncryptMasterKey)))
		authKey = bytes.Clone(b.Get([]byte(bucketKeyAuthKey)))
		return nil
	})
	return
}

func WriteBolt(salt, encryptMasterKey, authKey []byte) error {
	db, err := openBolt()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(bucketNameSpan))

		err = b.Put([]byte(bucketKeySalt), salt)
		if err != nil {
			return errors.Wrapf(err, "写数据库失败, key: %s", bucketKeySalt)
		}
		err = b.Put([]byte(bucketKeyEncryptMasterKey), encryptMasterKey)
		if err != nil {
			return errors.Wrapf(err, "写数据库失败, key: %s", bucketKeyEncryptMasterKey)
		}
		err = b.Put([]byte(bucketKeyAuthKey), authKey)
		if err != nil {
			return errors.Wrapf(err, "写数据库失败, key: %s", bucketKeyAuthKey)
		}

		return nil
	})
	return err
}
