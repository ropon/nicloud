package vmapis

import (
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
  "nicloud/utils"
  "nicloud/vm"
  "nicloud/vmerror"
  "strconv"
  "sync"
)

func Vnc(c *gin.Context)  {
  uuid := c.Query("uuid")
  host := c.Query("host")
  res := fmt.Sprintf("http://%s/novnc/vnc.html?path=websockify/?token=%s", host, uuid)
  c.JSON(200, res)
}

  //xlsx := excelize.NewFile()
  //xlsx.SetCellValue("Sheet1", "A2", "asdas")
  //
  //c.Header("Content-Type", "application/octet-stream")
  //c.Header("Content-Disposition", "attachment; filename="+"Workbook.xlsx")
  //c.Header("Content-Transfer-Encoding", "binary")
  //_ = xlsx.Write(c.Writer)

func Search(c *gin.Context)  {
  ct := c.Query("content")
  vms, err := vm.SearchVm(ct)
  res := make(map[string]interface{})
  if err != nil {
    c.JSON(200, res)
  }

  res["res"] = vms
  c.JSON(200, res)
}

func GetVminfo(c *gin.Context) {
  res := make(map[string]interface{})
  uuid := c.Query("uuid")
  iplist, err := vm.GetVmByUuid(uuid)
  if err != nil {
    res["res"] = iplist
    res["err"] = err
    c.JSON(200, res)
    return
  }

  res["res"] = iplist
  res["err"] = nil
  c.JSON(200, res)
}

func GetVmStatus(c *gin.Context) {
  host := c.Query("host")
  uuid := c.Query("uuid")

  res := make(map[string]interface{})
  vmstate, err := vm.VmStatus(uuid, host)

  if err != nil {
    res["res"] = vmstate
  }

  res["res"] = vmstate
  c.JSON(200, res)
}

func Getvmlist(c *gin.Context) {
  res := make(map[string]interface{})
  start, err := strconv.Atoi(c.Query("start"))
  if err != nil {
    res["err"] = vmerror.Error{Message: "参数错误"}
    c.JSON(200, res)
    return
  }

  token := c.Request.Header.Get("token")
  userid, err := utils.ParseToken(token)
  if err != nil {
    res["err"] = vmerror.Error{Message: "认证失败"}
    c.JSON(200, res)
    return
  }
  offset := 15
	pagenumber, vmcount,  err := vm.Getpagenumber(userid, offset)
	if err != nil {
    res["res"] = nil
    res["err"] = err
    c.JSON(200, res)
    return
  }

  vmlist, err := vm.VmList(userid, start, offset)

	res["res"] = vmlist
  res["pagenumber"] = pagenumber
  res["vmcount"] = vmcount
  res["err"] = err

	c.JSON(200, res)
}

func MigrateVmlive(c *gin.Context) {
  uuid := c.Query("uuid")
  migratehost := c.Query("migratehost")

  vmlist := vm.MigrateVmlive(uuid, migratehost)
  res := make(map[string]interface{})
  res["res"] = vmlist

  c.JSON(200, res)
}

func MigrateVm(c *gin.Context) {
  uuid := c.Query("uuid")
  migratehost := c.Query("migratehost")

  vmlist := vm.MigrateVm(uuid, migratehost)
  res := make(map[string]interface{})
  res["res"] = vmlist

  c.JSON(200, res)
}

func Createvm(c *gin.Context) {
  res := make(map[string]interface{})
  ip := c.PostForm("ip")
  cpu, _ := strconv.Atoi(c.PostForm("cpu"))
  mem, _ := strconv.Atoi(c.PostForm("mem"))
  host := c.PostForm("host")
  os := c.PostForm("os")
  datacenter := c.PostForm("datacenter")
  storage := c.PostForm("storage")
  vlan := c.PostForm("vlan")
  comment := c.PostForm("comment")

  token := c.Request.Header.Get("token")
  user, err := utils.ParseToken(token)

  if err != nil {
    res["err"] = vmerror.Error{Message: "认证失败"}
    c.JSON(200, res)
    return
  }

  v := vm.Vms{
    Ip: ip,
    Cpu: cpu,
    Mem: mem,
    Host: host,
    Os: os,
    Datacenter: datacenter,
    Storage: storage,
    Owner: user,
    Comment: comment,
  }

  validate := validator.New()
  err = validate.Struct(v)
  if err != nil {
    res["err"] = vmerror.Error{Message: "参数错误"}
    c.JSON(400, res)
    return
  }
  var rwLock sync.RWMutex
  rwLock.Lock()
  err = vm.Create(datacenter, storage, vlan, cpu, mem, ip, host, os, user, comment)
  rwLock.Unlock()
  res["err"] = err
  c.JSON(200, res)
}

func Addcomment(c *gin.Context) {
  uuid := c.Query("uuid")
  comment := c.Query("comment")
  res := make(map[string]interface{})
  r, err := vm.Updatecomments(uuid, comment)

  res["res"] = r
  res["err"] = err
  c.JSON(200, res)
}

func GetFlavor(c *gin.Context) {
	res := make(map[string]interface{})
	s, err := vm.Flavor()
	res["res"] = s
	res["err"] = err
	if err != nil {
		c.JSON(200, res)
	}

	c.JSON(200, res)
}

func Changeconfig(c *gin.Context) {
  id := c.Query("uuid")
  host := c.Query("host")
  vmhost := c.Query("vmhost")
  cpu, err := strconv.Atoi(c.Query("cpu"))
  if err != nil {
    c.JSON(400, vmerror.Error{Message: "参数错误"})
    return
  }

  oldcpu, err := strconv.Atoi(c.Query("oldcpu"))
  if err != nil {
    c.JSON(400, vmerror.Error{Message: "参数错误"})
    return
  }

  oldmem, err := strconv.Atoi(c.Query("oldmem"))
  if err != nil {
    c.JSON(400, vmerror.Error{Message: "参数错误"})
    return
  }

  mem, err := strconv.Atoi(c.Query("mem"))
  if err != nil {
    c.JSON(400, vmerror.Error{Message: "参数错误"})
    return
  }
  res := make(map[string]interface{})
  err = vm.Changeconfig(id, host, cpu, oldcpu, mem, oldmem, vmhost)
  res["err"] = err
  c.JSON(200, res)
}

func DeleteVM(c *gin.Context) {

	uuid := c.Query("uuid")
  datacenter := c.Query("datacenter")
  storage := c.Query("storage")

	res := make(map[string]interface{})
  var rwLock sync.RWMutex
	rwLock.Lock()
	err := vm.Delete(uuid, datacenter, storage)
	rwLock.Unlock()

	res["err"] = err
	c.JSON(200, res)
}

func Operation(c *gin.Context) {
	uuid := c.Query("uuid")
	host := c.Query("host")
	res := make(map[string]interface{})

	var err error

	o, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, res)
	}

	var s *vm.Vms
	switch o {
	case 0:
		err = vm.Shutdown(uuid, host)
  case 1:
    err = vm.Destroy(uuid, host)
	case 2:
		err = vm.Start(uuid, host)
  case 3:
    err = vm.PauseVm(uuid, host)
  case 4:
    err = vm.Reboot(uuid, host)
	}

	res["res"] = s
	res["err"] = err
	c.JSON(200, res)
}

func Rebuild(c *gin.Context)  {
  uuid := c.Query("uuid")
  datacenter := c.Query("datacenter")
  storage := c.Query("storage")
  osname := c.Query("osname")
  host := c.Query("host")

  res := make(map[string]interface{})
  err := vm.Rebuildimg(osname, storage, datacenter, uuid, host)
  res["err"] = err
  c.JSON(200, res)
}

func Createsnap(c *gin.Context)  {
  uuid := c.PostForm("uuid")
  datacenter := c.PostForm("datacenter")
  storage := c.PostForm("storage")
  snapname := c.PostForm("snapname")

  res := make(map[string]interface{})
  err := vm.Creatsnap(uuid, datacenter, storage, snapname)
  res["err"] = err
  c.JSON(200, res)
}

func Getsnap(c *gin.Context)  {
  uuid := c.Query("uuid")
  datacenter := c.Query("datacenter")
  storage := c.Query("storage")

  res := make(map[string]interface{})
  s, err := vm.Getsnap(datacenter, storage, uuid)
  res["res"] = s
  res["err"] = err
  c.JSON(200, res)
}

func Rollback(c *gin.Context)  {
  uuid := c.Query("uuid")
  datacenter := c.Query("datacenter")
  storage := c.Query("storage")
  snapname := c.Query("snapname")

  res := make(map[string]interface{})
  err := vm.RollbackSnap(uuid, snapname,  datacenter, storage)

  res["err"] = err
  c.JSON(200, res)
}
