import { Outlet, useNavigate } from 'react-router-dom'
import { Button, Layout, Space, Typography, theme } from 'antd'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { useAuth } from '../auth/AuthProvider'
import { graphqlSdk } from '../graphql/client'
import ThemeSwitcher from './ThemeSwitcher'

const { Header, Content, Sider } = Layout
const { Text } = Typography

export default function AppLayout() {
  const navigate = useNavigate()
  const { token } = theme.useToken()
  const auth = useAuth()

  const handleLogout = async () => {
    try {
      await graphqlSdk(auth.token).Logout()
    } catch {
      // Ошибку logout на сервере не показываем: локальная сессия всё равно закрывается
    } finally {
      auth.logout()
      navigate('/login')
    }
  }

  return (
    <Layout className="app-shell">
      <Sider width={220} theme="dark">
        <div className="brand">
          <UserOutlined />
          <span>Frontend</span>
        </div>
      </Sider>
      <Layout>
        <Header
          className="app-header"
          style={{
            background: token.colorBgContainer,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <Space>
            <Text>{auth.user?.nickname || auth.user?.email}</Text>
            <ThemeSwitcher />
            <Button icon={<LogoutOutlined />} type="text" onClick={handleLogout}>
              Выйти
            </Button>
          </Space>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
