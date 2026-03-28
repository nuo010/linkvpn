<template>
  <div class="config-page">
    <div class="card alert-warning">
      <span class="alert-icon">⚠️</span>
      常用参数建议通过表单配置；写入 `server.conf` 后仍需重启 OpenVPN 服务才能生效。
    </div>

    <div class="card config-card">
      <h3 class="section-title">VPN 服务配置</h3>
      <p class="section-desc">通过表单完成全部常用配置，下方仅展示最终生成的 `server.conf` 结果，不再允许直接编辑。</p>
      <el-form label-position="top" class="params-form">
        <div class="group-title">基础网络</div>
        <div class="params-grid">
          <el-form-item label="协议">
            <el-select v-model="form.protocol">
              <el-option label="UDP" value="udp" />
              <el-option label="TCP" value="tcp" />
            </el-select>
          </el-form-item>

          <el-form-item label="设备类型">
            <el-select v-model="form.device">
              <el-option label="TUN" value="tun" />
              <el-option label="TAP" value="tap" />
            </el-select>
          </el-form-item>

          <el-form-item label="拓扑模式">
            <el-select v-model="form.topology">
              <el-option label="subnet" value="subnet" />
              <el-option label="net30" value="net30" />
            </el-select>
          </el-form-item>

          <el-form-item label="端口">
            <el-input v-model="form.port" type="number" min="1" max="65535" />
          </el-form-item>

          <el-form-item label="虚拟网段">
            <el-input v-model="form.subnet" placeholder="10.8.8.0/24" />
          </el-form-item>

          <el-form-item label="最大连接数">
            <el-input v-model="form.max_connections" type="number" min="1" placeholder="200" />
          </el-form-item>

          <el-form-item label="保活参数（keepalive）">
            <el-input v-model="form.keepalive" placeholder="10 120" />
            <div class="field-help">对应 `keepalive 10 120`。前一个数表示心跳间隔秒数，后一个数表示超时重启秒数。</div>
          </el-form-item>
        </div>

        <div class="group-title">推送与路由</div>
        <div class="params-grid">
          <el-form-item label="推送 DNS 1">
            <el-input v-model="form.push_dns1" placeholder="8.8.8.8" />
          </el-form-item>

          <el-form-item label="推送 DNS 2">
            <el-input v-model="form.push_dns2" placeholder="1.1.1.1" />
          </el-form-item>

          <el-form-item label="IPv6 网段">
            <el-input v-model="form.ipv6_subnet" :disabled="!form.ipv6" placeholder="fd00:8::/64" />
          </el-form-item>
        </div>

        <el-form-item label="推送路由">
          <el-input
            v-model="form.push_routes"
            type="textarea"
            :rows="4"
            placeholder="每行一个 CIDR，例如&#10;192.168.10.0/24&#10;172.16.0.0/16"
          />
        </el-form-item>

        <div class="switch-row">
          <el-checkbox v-model="form.vpn_gateway">推送默认路由（redirect-gateway）</el-checkbox>
          <el-checkbox v-model="form.client_to_client">允许客户端互访（client-to-client）</el-checkbox>
          <el-checkbox v-model="form.ipv6">启用 IPv6</el-checkbox>
        </div>

        <div class="group-title">安全与运行</div>
        <div class="params-grid">
          <el-form-item label="Cipher">
            <el-select v-model="form.cipher">
              <el-option label="AES-256-GCM" value="AES-256-GCM" />
              <el-option label="AES-128-GCM" value="AES-128-GCM" />
              <el-option label="AES-256-CBC" value="AES-256-CBC" />
            </el-select>
          </el-form-item>

          <el-form-item label="Auth">
            <el-select v-model="form.auth">
              <el-option label="SHA256" value="SHA256" />
              <el-option label="SHA1" value="SHA1" />
              <el-option label="SHA512" value="SHA512" />
            </el-select>
          </el-form-item>

          <el-form-item label="运行用户">
            <el-input v-model="form.run_user" placeholder="nobody" />
          </el-form-item>

          <el-form-item label="运行组">
            <el-input v-model="form.run_group" placeholder="nogroup" />
          </el-form-item>

          <el-form-item label="日志级别">
            <el-select v-model="form.verb">
              <el-option label="1" value="1" />
              <el-option label="2" value="2" />
              <el-option label="3" value="3" />
              <el-option label="4" value="4" />
            </el-select>
          </el-form-item>
        </div>

        <div class="switch-row">
          <el-checkbox v-model="form.persist_key">persist-key</el-checkbox>
          <el-checkbox v-model="form.persist_tun">persist-tun</el-checkbox>
          <el-checkbox v-model="form.explicit_exit_notify">explicit-exit-notify</el-checkbox>
        </div>
        <div class="field-help inline-help">
          `persist-key`：重连时尽量保留密钥状态；`persist-tun`：重连时尽量保留虚拟网卡；`explicit-exit-notify`：UDP 断开时主动通知服务端。
        </div>
        <div class="field-help inline-help">点击“保存表单配置”时会默认同时写入 `server.conf`。</div>
      </el-form>

      <div class="config-actions">
        <el-button type="primary" @click="saveParams">保存表单配置</el-button>
        <el-button @click="loadParams">重新加载表单</el-button>
        <el-button @click="resetParams">恢复默认参数</el-button>
        <el-button
          type="warning"
          plain
          class="restart-service-btn"
          :loading="restartLoading"
          @click="restartService"
        >
          <span class="restart-icon">↻</span>
          {{ restartLoading ? '正在重启…' : '重启 OpenVPN' }}
        </el-button>
      </div>
    </div>

    <div class="card config-card">
      <h3 class="section-title">生成后的配置预览</h3>
      <p class="section-desc">这里仅用于查看最终生成的 `server.conf`，如需修改，请回到上方表单。</p>
      <div class="preview-actions">
        <el-button @click="loadFile">刷新预览</el-button>
      </div>
      <pre class="config-preview">{{ fileContent }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '../api/client'

const fileContent = ref('')
const restartLoading = ref(false)
const form = ref({
  port: '1194',
  protocol: 'udp',
  device: 'tun',
  topology: 'subnet',
  max_connections: '200',
  subnet: '10.8.8.0/24',
  push_routes: '',
  push_dns1: '8.8.8.8',
  push_dns2: '2001:4860:4860::8888',
  vpn_gateway: false,
  client_to_client: true,
  ipv6: false,
  ipv6_subnet: 'fd00:8::/64',
  keepalive: '10 120',
  cipher: 'AES-256-GCM',
  auth: 'SHA256',
  run_user: 'nobody',
  run_group: 'nogroup',
  persist_key: true,
  persist_tun: true,
  verb: '3',
  explicit_exit_notify: true,
  auto_apply_to_config: true,
})

function normalizeParams(data = {}) {
  return {
    port: String(data.port ?? 1194),
    protocol: data.protocol === 'tcp' ? 'tcp' : 'udp',
    device: data.device === 'tap' ? 'tap' : 'tun',
    topology: data.topology === 'net30' ? 'net30' : 'subnet',
    max_connections: String(data.max_connections ?? 200),
    subnet: data.subnet || '10.8.8.0/24',
    push_routes: data.push_routes || '',
    push_dns1: data.push_dns1 || '',
    push_dns2: data.push_dns2 || '',
    vpn_gateway: !!data.vpn_gateway,
    client_to_client: data.client_to_client !== false,
    ipv6: !!data.ipv6,
    ipv6_subnet: data.ipv6_subnet || 'fd00:8::/64',
    keepalive: data.keepalive || '10 120',
    cipher: data.cipher || 'AES-256-GCM',
    auth: data.auth || 'SHA256',
    run_user: data.run_user || 'nobody',
    run_group: data.run_group || 'nogroup',
    persist_key: data.persist_key !== false,
    persist_tun: data.persist_tun !== false,
    verb: String(data.verb || '3'),
    explicit_exit_notify: data.explicit_exit_notify !== false,
    auto_apply_to_config: data.auto_apply_to_config !== false,
  }
}

function resetParams() {
  form.value = normalizeParams()
  ElMessage.success('已恢复默认参数，保存后才会写入服务器')
}

async function loadParams() {
  try {
    const { data } = await client.get('/config/params')
    form.value = normalizeParams(data)
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载表单配置失败')
  }
}

async function loadFile() {
  try {
    const { data } = await client.get('/config/file')
    fileContent.value = data.content || ''
    if (!fileContent.value.trim()) {
      ElMessage.warning('当前无配置文件，已填充默认内容，保存后生效。')
    }
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '加载失败')
  }
}

async function saveParams() {
  try {
    const saveResp = await client.post('/config/params', { ...form.value, auto_apply_to_config: false })
    const applyResp = await client.post('/config/params/apply')
    await loadFile()
    const message = applyResp.data?.message || saveResp.data?.message || '表单配置已保存并写入 server.conf'
    ElMessage.success(message)
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '保存失败')
  }
}

async function restartService() {
  restartLoading.value = true
  try {
    const { data } = await client.post('/config/restart')
    ElMessage.success(data.message || '重启已触发')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '重启失败')
  } finally {
    restartLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadParams(), loadFile()])
})
</script>

<style scoped>
.config-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.alert-warning {
  background: #fdf6ec;
  border-color: #f5dab1;
  color: #b88230;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.alert-icon {
  font-size: 1.2rem;
}
.config-card {
  overflow: hidden;
}
.section-title {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.section-desc {
  margin: 0 0 1rem;
  color: var(--muted);
  font-size: 0.9rem;
}
.params-form {
  margin-bottom: 0.5rem;
}
.params-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0 1rem;
}
.group-title {
  margin: 0 0 0.85rem;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text);
}
.switch-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 0.25rem;
  margin-bottom: 0.5rem;
}
.field-help {
  margin-top: 0.35rem;
  color: var(--muted);
  font-size: 0.82rem;
  line-height: 1.5;
}
.inline-help {
  margin-top: 0.1rem;
  margin-bottom: 0.5rem;
}
.config-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
/* 重启：醒目但不抢眼，与主按钮区分 */
.restart-service-btn {
  margin-left: auto;
  font-weight: 600;
  border-width: 2px;
  box-shadow: 0 2px 8px rgba(230, 162, 60, 0.25);
}
.restart-service-btn:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(230, 162, 60, 0.35);
}
.restart-icon {
  display: inline-block;
  margin-right: 6px;
  font-size: 1.1em;
  line-height: 1;
  vertical-align: -0.05em;
}
.preview-actions {
  margin-bottom: 0.75rem;
}
.config-preview {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 0.85rem;
  min-height: 320px;
  width: 100%;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.75rem 0.9rem;
  line-height: 1.6;
  background: #f8fafc;
  white-space: pre-wrap;
  word-break: break-word;
  overflow: auto;
}

@media (max-width: 960px) {
  .restart-service-btn {
    margin-left: 0;
  }
}
</style>
