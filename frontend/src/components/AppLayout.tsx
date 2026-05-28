import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Button, Checkbox, Layout, Modal, Space, Typography } from "antd";
import { LogoutOutlined, UserOutlined } from "@ant-design/icons";
import { useEffect, useState } from "react";
import { useAuth } from "../auth/authStore";
import { graphqlSdk } from "../graphql/client";
import { isSessionInvalidError } from "../graphql/errors";
import ThemeSwitcher from "./ThemeSwitcher";

const { Header, Content, Sider } = Layout;
const { Text } = Typography;

export default function AppLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const auth = useAuth();
  const [logoutOpen, setLogoutOpen] = useState(false);
  const [logoutEverywhere, setLogoutEverywhere] = useState(false);

  const profileQuery = useQuery({
    queryKey: ["my-profile", auth.user?.id],
    queryFn: () =>
      graphqlSdk(auth.token)
        .MyProfile()
        .then((data) => data.myProfile),
    enabled: Boolean(auth.token),
    retry: (failureCount, error) =>
      !isSessionInvalidError(error) && failureCount < 3,
  });

  const displayName =
    profileQuery.data?.nickname || auth.user?.nickname || auth.user?.email;

  useEffect(() => {
    if (profileQuery.isError && isSessionInvalidError(profileQuery.error)) {
      const next = encodeURIComponent(location.pathname + location.search);
      auth.logout();
      navigate(`/login?next=${next}`, { replace: true });
    }
  }, [
    auth,
    location.pathname,
    location.search,
    navigate,
    profileQuery.error,
    profileQuery.isError,
  ]);

  const handleLogout = async () => {
    try {
      await graphqlSdk(auth.token).Logout({ allDevices: logoutEverywhere });
    } catch {
      // Ошибку logout на сервере не показываем: локальная сессия всё равно закрывается
    } finally {
      auth.logout();
      navigate("/login");
    }
  };

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
            <Text>{displayName}</Text>
            <ThemeSwitcher />
            <Button
              icon={<LogoutOutlined />}
              type="text"
              onClick={() => setLogoutOpen(true)}
            >
              Выйти
            </Button>
          </Space>
        </Header>
        <Content className="m-6 min-h-[280px]">
          <Outlet />
        </Content>
      </Layout>
      <Modal
        okText="Выйти"
        open={logoutOpen}
        title="Выйти из системы"
        width={360}
        onCancel={() => setLogoutOpen(false)}
        onOk={handleLogout}
      >
        <Checkbox
          checked={logoutEverywhere}
          onChange={(event) => setLogoutEverywhere(event.target.checked)}
        >
          на всех устройствах
        </Checkbox>
      </Modal>
    </Layout>
  );
}
