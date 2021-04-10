package vmapis

import (
  "github.com/gin-gonic/gin"
  "goblog/vmcommon"
  "strconv"
)

func Getvmlist(c *gin.Context) {

  vmlist := vmcommon.VmList()
  res := make(map[string]interface{})
  res["res"] = vmlist

  c.JSON(200, res)
}

func Createvm(c *gin.Context) {
  cpu, _  := strconv.Atoi(c.Query("cpu"))
  mem, _ := strconv.Atoi(c.Query("mem"))

  create, err := vmcommon.Create(cpu, mem)
  if err != nil {
    c.JSON(500, err)
  }
  res := make(map[string]interface{})
  res["res"] = create

  c.JSON(200, res)
}

func GetStatus(c *gin.Context) {
  s, _ := vmcommon.VmStatus("31a803b2-5f11-4f14-875f-d14347db13fb")
  res := make(map[string]interface{})
  res["res"] = s
  c.JSON(200, res)
}

func  Operation(c *gin.Context)  {
  uuid := c.Query("uuid")
  res := make(map[string]interface{})

  o, err := strconv.Atoi(c.Param("id"))
  if err != nil {
    c.JSON(400, res)
  }

  var s *vmcommon.Vms
  switch o {
  case 1: s, _ = vmcommon.Start(uuid)
  case 0: s, _ = vmcommon.Shutdown(uuid)
  }

  res["res"] = s
  res["err"] = err
  c.JSON(200, res)
}
