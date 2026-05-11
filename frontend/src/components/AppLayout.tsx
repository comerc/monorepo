import { Outlet, useNavigate } from 'react-router-dom'
import { Button, Layout, Space, Typography } from 'antd'
import { LogoutOutlined, UserOutlined } from '@ant-design/icons'
import { useAuth } from '../auth/AuthProvider'
import { graphqlSdk } from '../graphql/client'
import ThemeSwitcher from './ThemeSwitcher'

const { Header, Content, Sider } = Layout
const { Text } = Typography

export default function AppLayout() {
  const navigate = useNavigate()
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
    <Layout className="min-h-screen">
      <Sider width={220} theme="dark">
        <div className="flex h-16 items-center justify-center gap-2 text-lg font-semibold text-white">
          <UserOutlined />
          <span>Frontend</span>
        </div>
      </Sider>
      <Layout>
        <Header className="flex items-center justify-end border-b border-app-border bg-app-header px-6 dark:border-app-border-dark dark:bg-app-header-dark">
          <Space>
            <Text>{auth.user?.nickname || auth.user?.email}</Text>
            <ThemeSwitcher />
            <Button icon={<LogoutOutlined />} type="text" onClick={handleLogout}>
              Выйти
            </Button>
          </Space>
        </Header>
        <Content className="m-6 min-h-[280px]">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
