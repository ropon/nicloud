<template>
  <div>
    <div class="col-sm-12 form-group" style="border-bottom: 1px green solid">
      <h4>修改配置</h4>
    </div>

    <div class="col-sm-8 col-sm-offset-2 choose" style="margin-top:30px;">
      <ul class="nav nav-pills nav-stacked">
        <li><strong>uuid</strong>:{{ uuid }}</li>
        <li><strong>ip</strong>:{{ ip }}</li>
        <li><strong>os</strong>:{{ os }}</li>
        <li><strong>host</strong>:{{ host }}</li>
        <li><strong>cpu</strong>:{{ cpu }}核</li>
        <li><strong>mem</strong>:{{ mem }}G</li>
        <li><strong>owner</strong>:{{ owner }}</li>
        <li><strong>comment</strong>:{{ comment }}</li>
      </ul>
    </div>
    <div class="col-sm-8 col-sm-offset-2 choose" style="margin-top:30px; margin-bottom:30px">
      <div class="col-sm-8">
        <div class="col-sm-2" style="margin-top:1px">选择配置</div>
        <div class="col-sm-9" style="padding-left:0">
          <select class="col-sm-3" v-model="configvalue">
            <option v-for="c in config" :value="c" :key="c.id"> {{ c.Cpu }}核/{{ c.Mem }}G </option>
          </select>
          <button @click="changeconfig" style="margin-left:40px;margin-top:1px" type="button" class="btn btn-success btn-xs">提交</button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      uuid: '',
      ip: '',
      os: '',
      host: '',
      cpu: '',
      mem: '',
      owner: '',
      comment: '',
      configvalue: {},
      config: []
    }
  },

  created: function() {
    this.vminfo()
    this.getflavor()
  },

  methods: {
    getflavor: function() {
      var apiurl = `/api/vm/getflavor`
      this.$http.get(apiurl).then(response => {
        this.config = response.data.res
      })
    },

    changeconfig: function() {
      var apiurl = `/api/vm/changeconfig`
      this.$http
        .get(apiurl, { params: { uuid: this.uuid, host: this.host, cpu: this.configvalue.Cpu, oldcpu: this.cpu, oldmem: this.mem, mem: this.configvalue.Mem, vmhost: this.host } })
        .then(response => {
          if (response.data.err === null) {
            alert('修改成功')
          } else {
            alert("修改失败'(" + response.data.err.Message + "')")
          }
        })
    },
    create: function() {
      if (typeof this.startip === 'undefined' || this.startip === null || this.startip === '' || typeof this.endip === 'undefined' || this.endip === null || this.endip === '') {
        alert('输入为空')
        return
      }

      var apiurl = `/api/networks/createip`
      this.$http.get(apiurl, { params: { startip: this.startip, endip: this.endip, vlan: this.vlan } }).then(response => {
        if (response.data.res === null) {
          alert('创建成功')
        } else {
          alert("创建失败'(" + response.data.res.Message + "')")
        }
      })
    },

    vminfo: function() {
      var v = this.$store.state.editsetting.uuid
      if (v === null || typeof v === 'undefined' || v === '' || v === 'undefined') {
        this.uuid = sessionStorage.getItem('uuid')
        this.ip = sessionStorage.getItem('ip')
        this.os = sessionStorage.getItem('os')
        this.host = sessionStorage.getItem('host')
        this.cpu = sessionStorage.getItem('cpu')
        this.mem = sessionStorage.getItem('mem')
        this.owner = sessionStorage.getItem('owner')
        this.comment = sessionStorage.getItem('comment')
      } else {
        this.uuid = this.$store.state.editsetting.uuid
        this.ip = this.$store.state.editsetting.ip
        this.os = this.$store.state.editsetting.os
        this.host = this.$store.state.editsetting.host
        this.cpu = this.$store.state.editsetting.cpu
        this.mem = this.$store.state.editsetting.mem
        this.owner = this.$store.state.editsetting.owner
        this.comment = this.$store.state.editsetting.comment
        sessionStorage.setItem('uuid', this.$store.state.editsetting.uuid)
        sessionStorage.setItem('ip', this.$store.state.editsetting.ip)
        sessionStorage.setItem('os', this.$store.state.editsetting.os)
        sessionStorage.setItem('host', this.$store.state.editsetting.host)
        sessionStorage.setItem('cpu', this.$store.state.editsetting.cpu)
        sessionStorage.setItem('mem', this.$store.state.editsetting.mem)
        sessionStorage.setItem('owner', this.$store.state.editsetting.owner)
        sessionStorage.setItem('comment', this.$store.state.editsetting.comment)
      }
    }
  }
}
</script>
<style scoped>
.createip {
  font-weight: 500;
}

.vlaninfo {
  font-weight: 501;
}
.col-sm-2 {
  padding-left: 0;
}

.choose {
  padding: 10px;
  border-style: solid;
  border-color: #ddd;
  border-width: 1px;
  border-radius: 4px 4px 0 0;
}

.col-sm-6 {
  padding: 10px;
  border-style: solid;
  border-color: #ddd;
  border-width: 1px;
  border-radius: 4px 4px 0 0;
}

.startip {
  margin-top: 10px;
  padding-right: 0px;
}

.endip {
  margin-top: 10px;
  padding-right: 0px;
  padding-left: 0px;
}

.col-sm-4 label {
  float: right;
}
select {
  font-family: '微软雅黑';
  border: 1px #ccc solid;
  border-radius: 5px;
}

.details-content .article-cont p {
  padding: 30px 0 0 5px;
}

label {
  font-weight: 400;
  margin-top: 5px;
}
</style>
