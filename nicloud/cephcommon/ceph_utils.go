package cephcommon

import (
	"fmt"
	"github.com/ceph/go-ceph/rados"
	rbd "github.com/ceph/go-ceph/rbd"
	"nicloud/dbs"
	"nicloud/vmerror"
	"time"
)

type Vms_Ceph struct {
	Uuid        string `gorm:"primary_key" json:"Uuid" validate:"required"`
	Name        string `gorm:"unique" json:"Name" validate:"required"`
	Pool        string `json:"Pool" validate:"required"`
	Contain     string
	Remainder   string
	Datacenter  string `json:"Datacenter" validate:"required"`
	Ceph_secret string `json:"Ceph_secret" validate:"required"`
	Ips         string `json:"Ips" validate:"required"`
	Port        string `json:"Port" validate:"required"`
	Comment     string
	Status      int8
}

func Delete(uuid string) error {
	dbs, err := db.NicloudDb()
	if err != nil {
		return err
	}

	dberr := dbs.Where("uuid=?", uuid).Delete(Vms_Ceph{})
	if dberr.Error != nil {
		return dberr.Error
	}

	return nil
}

func (ceph Vms_Ceph) Add(uuid string, name string, pool string, datacenter string, ceph_secret string, ips string, port string, comment string) error {
	c := &Vms_Ceph{
		Uuid:        uuid,
		Name:        name,
		Pool:        pool,
		Datacenter:  datacenter,
		Ceph_secret: ceph_secret,
		Ips:         ips,
		Port:        port,
		Comment:     comment,
		Status:      1,
	}
	dbs, err := db.NicloudDb()
	if err != nil {
		return err
	}
	errdb := dbs.Debug().Create(c)
	if errdb.Error != nil {
		return errdb.Error
	}
	return nil
}

func (ceph Vms_Ceph) Get() ([]*Vms_Ceph, error) {
	conn, err := CephConn()
	if err != nil {
		return nil, err
	}

	clusterstat, err := conn.GetClusterStats()
	if err != nil {
		return nil, err
	}

	dbs, err := db.NicloudDb()
	if err != nil {
		return nil, err
	}
	c := []*Vms_Ceph{}
	dbs.Find(&c)
	for _, v := range c {
		v.Contain = fmt.Sprintf("%.2f", float64(clusterstat.Kb)/float64(1024*1024*1024))
		v.Remainder = fmt.Sprintf("%.2f", float64(clusterstat.Kb_used)/float64(1024*1024*1024))
	}
	return c, nil
}

//func (ceph Vms_Ceph)IncreaseContain(uuid string, contain int) error {
//  dbs, err := db.NicloudDb()
//  if err != nil {
//    return err
//  }
//
//  cephinfo, err := ceph.Cephinfobyuuid(uuid)
//  if err != nil {
//    return err
//  }
//
//  if contain+cephinfo.Remainder > cephinfo.Contain {
//    return vmerror.Error{Message: "数据池容量不足"}
//  }
//
//  errdb := dbs.Model(ceph).Where("uuid=?", uuid).Update("remainder", contain+cephinfo.Remainder)
//  if errdb.Error != nil {
//    return err
//  }
//  return nil
//}

func Getpool(datacenter string, storage string) ([]*Vms_Ceph, error) {
	dbs, err := db.NicloudDb()
	if err != nil {
		return nil, err
	}
	c := []*Vms_Ceph{}
	dbs.Where("datacenter=? and uuid=?", datacenter, storage).Find(&c)
	return c, nil
}

func (ceph Vms_Ceph) Cephinfobyuuid(uuid string) (*Vms_Ceph, error) {
	dbs, err := db.NicloudDb()
	if err != nil {
		return nil, err
	}
	c := &Vms_Ceph{}
	errdb := dbs.Where("uuid=?", uuid).First(c)
	if errdb.Error != nil {
		return nil, err
	}
	return c, nil
}

func CephConn() (*rados.Conn, error) {
	conn, err := rados.NewConn()

	if err != nil {
		return nil, err
	}
	err = conn.ReadDefaultConfigFile() // /etc/ceph/ceph.conf
	if err != nil {
		return nil, err
	}
	err = conn.Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c Vms_Ceph) rename(img *rbd.Image, blockname string) error {
	err := img.Rename(blockname)
	if err != nil {
		return err
	}
	return nil
}

func (c Vms_Ceph) RenameBlock(uuid string, desname string) error {
	img, err := c.Getimgbyname(uuid, c.Pool)
	if err != nil {
		return err
	}
	err = c.rename(img, desname)
	if err != nil {
		return err
	}
	return nil
}

func (c Vms_Ceph) Rm_image(uuid string, pool string) (string, error) {
	ioctx, err := c.Ceph_ioctx(pool)
	if err != nil {
		return "", err
	}

	img := rbd.GetImage(ioctx, uuid)
	Archive_img := "x_" + (time.Now().Format("20060102150405")) + uuid
	err = c.rename(img, Archive_img)
	if err != nil {
		return "", err
	}

	return Archive_img, nil
}

func image_ctx(ctx *rados.IOContext, cephblock string) *rbd.Image {
	imagectx := rbd.GetImage(ctx, cephblock)
	return imagectx
}

func (c Vms_Ceph) RbdClone(id string, cephblock string, snap string, pool string) (string, error) {
	ioctx, err := c.Ceph_ioctx(pool)

	//openimage
	o, err := rbd.OpenImage(ioctx, cephblock, snap)
	if err != nil {
		return "", err
	}

	img_ctx := image_ctx(ioctx, cephblock)

	//保护快照 important
	snapshot := o.GetSnapshot(snap)
	snapisprotected, _ := snapshot.IsProtected()

	if snapisprotected == false {
		return "", vmerror.Error{Message: "镜像快照保护未设置"}
	}

	_, err = img_ctx.Clone(snap, ioctx, id, rbd.FeatureLayering, 12)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (ceph Vms_Ceph) Createcephblock(uuid string, contain int, pool string) error {
	ioctx, err := ceph.Ceph_ioctx(pool)
	if err != nil {
		return err
	}

	_, err = rbd.Create(ioctx, uuid, uint64(1024*1024*1024*contain), 0)
	if err != nil {
		return err
	}
	return nil
}

func (c Vms_Ceph) Ceph_ioctx(pool string) (*rados.IOContext, error) {
	conn, err := CephConn()
	if err != nil {
		return nil, vmerror.Error{Message: "ceph 连接失败"}
	}

	ioctx, err := conn.OpenIOContext(pool)
	if err != nil {
		return nil, vmerror.Error{Message: "获取ceph池句柄失败"}
	}
	return ioctx, nil
}

func (c Vms_Ceph) Changename(uuid string, cephblock string, snap string, pool string, oldname string) error {
	ioctx, err := c.Ceph_ioctx(pool)
	if err != nil {
		return err
	}
	img, err := c.RbdClone(uuid, cephblock, snap, pool)
	if err != nil {
		return err
	}
	_, err = c.Rm_image(oldname, pool)
	if err != nil {
		return err
	}

	fd := image_ctx(ioctx, img)
	err = c.rename(fd, oldname)
	if err != nil {
		return err
	}
	return nil
}

type Vms_snaps struct {
	Id          int `gorm:"primary_key;AUTO_INCREMENT"`
	Vm_uuid     string
	Datacenter  string
	Storage     string
	Snap        string
	Create_time time.Time `json:"Create_time"`
	Status      bool
}

func (c Vms_Ceph) Getimgbyname(imgname string, pool string) (*rbd.Image, error) {
	ctx, err := c.Ceph_ioctx(pool)
	if err != nil {
		return nil, err
	}

	errimg, err := rbd.OpenImage(ctx, imgname, "")
	if err != nil {
		return nil, err
	}
	return errimg, nil
}

func (c Vms_Ceph) Createimgsnap(vmid string, snapname string, pool string) error {
	img, err := c.Getimgbyname(vmid, pool)
	if err != nil {
		return err
	}

	_, err = img.CreateSnapshot(snapname)
	if err != nil {
		return err
	}

	return nil
}

func (c Vms_Ceph) Rollback(vmid string, snapname string, pool string) error {
	img, err := c.Getimgbyname(vmid, pool)
	if err != nil {
		return err
	}

	s := img.GetSnapshot(snapname)
	err = s.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func (c Vms_Ceph) Delsnap(imageid string, snapname string, pool string) error {
	img, err := c.Getimgbyname(imageid, pool)
	if err != nil {
		return err
	}

	s := img.GetSnapshot(snapname)

	b, err := s.IsProtected()
	if err != nil {
		return err
	}

	if b {
		err = s.Unprotect()
		if err != nil {
			return err
		}
	}
	err = s.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (c Vms_Ceph) CreateSnapAndProtect(pool string, imgid string) (string, error) {
	img, err := c.Getimgbyname(imgid, pool)
	if err != nil {
		return "", err
	}

	snapname := imgid + "-snap-" + time.Now().Format("200601021504")
	_, err = img.CreateSnapshot(snapname)
	if err != nil {
		return "", err
	}

	err = c.SnapProtect(imgid, pool, snapname)
	if err != nil {
		return "", err
	}
	return snapname, nil
}

func (c Vms_Ceph) SnapProtect(imgid string, pool string, snapname string) error {
	img, err := c.Getimgbyname(imgid, pool)
	if err != nil {
		return err
	}

	s := img.GetSnapshot(snapname)
	b, err := s.IsProtected()
	if err != nil {
		return err
	}

	if b == false {
		s.Protect()
	}

	return nil
}

func (c Vms_Ceph) ListChildernImages(storage string, imageid string) ([]string, error) {
	cephinfo, err := c.Cephinfobyuuid(storage)
	if err != nil {
		return nil, err
	}
	img, err := c.Getimgbyname(imageid, cephinfo.Pool)
	if err != nil {
		return nil, err
	}

	_, images, err := img.ListChildren()
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (c Vms_Ceph) Delimgpermanent(storage string, uuid string) error {
	cephinfo, err := c.Cephinfobyuuid(storage)
	if err != nil {
		return err
	}

	ctx, err := c.Ceph_ioctx(cephinfo.Pool)
	if err != nil {
		return err
	}

	err = rbd.TrashRemove(ctx, uuid, true)
	if len(err.Error()) > 0 {
		return err
	}
	return nil
}
