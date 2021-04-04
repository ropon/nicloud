package vmcommon

import (
  "fmt"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/mysql" //这个一定要引入哦！！
  libvirt "libvirt.org/libvirt-go"
)

type Vms struct {
  Uuid       string
  Name       string
  Cpu        int8
  Mem        int8
  Createtime string
  Owner      string
  Comment    string
  Status     string
}

func vmdb() *gorm.DB {
  db, errDb := gorm.Open("mysql", "modis:modis@(10.0.90.151:3306)/gocloud")
  if errDb != nil {
    fmt.Println(errDb)
  }
  return db
}


type Vm_xmls struct {
  Ostype string
  Osxml string
}

func libvirtconn() *libvirt.Connect {
  conn, err := libvirt.NewConnect("qemu:///system")
  if err != nil {
    fmt.Println(err)
  }
  return conn
}

func VmStatus(uuid string) (string, error) {
  var stats map[libvirt.DomainState]string
  stats = make(map[libvirt.DomainState]string)
  stats[5] = "关机"
  stats[1] = "运行"

  conn := libvirtconn()
  vm, err := conn.LookupDomainByUUIDString(uuid)

  if err != nil {
    return "vm not found", err
  }

  state, _ , err1  := vm.GetState()

  if err1 != nil {
    return "vm not found", err1
  }

  return stats[state], err1
}

func Shutdown(uuid string) (string, error) {
  /*start vm*/
  conn := libvirtconn()
  vm, err := conn.LookupDomainByUUIDString(uuid)
  fmt.Println(vm)
  if err != nil {
    fmt.Println(err)
  }

  err1 := vm.Shutdown()
  if err1 != nil {
    return "error", err1
  }
  s, err2 := VmStatus(uuid)
  if err2 != nil {
    fmt.Println(err2)
  }
  return s, err1
}

func Start(uuid string) (string, error) {
  /*start vm*/
  conn := libvirtconn()
  vm, err := conn.LookupDomainByUUIDString(uuid)

  if err != nil {
    fmt.Println(err)
  }

  err1 := vm.Create()
  if err1 != nil {
    fmt.Println(err1)
  }

  s, err2 := VmStatus(uuid)
  if err2 != nil {
    fmt.Println(err2)
  }
  return s, err2
}

func Create(uuid string) (bool, error) {
  /*create a vm*/
  db := vmdb()
  var x Vm_xmls
  db.First(&x, "ostype = ?", "linux")

  conn := libvirtconn()
  _, err1 := conn.DomainDefineXML(x.Osxml)

  if err1 != nil {
    return false, err1
  }
  return true, err1
}

func VmList() []*Vms {
  db := vmdb()
  var v []*Vms
  db.Find(&v)
  //s, _ := VmStatus("31a803b2-5f11-4f14-875f-d14347db13fb")
  //fmt.Println(s)
  for _, e := range(v) {
    s, _ := VmStatus(e.Uuid)
    e.Status = s
  }
  return v
}
