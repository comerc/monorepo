import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  App as AntdApp,
  Button,
  Card,
  Form,
  Input,
  Spin,
  Typography,
} from "antd";
import { useAuth } from "../auth/authStore";
import { graphqlSdk } from "../graphql/client";
import { isSessionInvalidError } from "../graphql/errors";

const { Text, Title } = Typography;

interface ProfileFormValues {
  nickname: string;
  email: string;
}

export default function ProfilePage() {
  const [form] = Form.useForm<ProfileFormValues>();
  const queryClient = useQueryClient();
  const auth = useAuth();
  const { message } = AntdApp.useApp();
  const profileQueryKey = ["my-profile", auth.user?.id];
  const watchedNickname = Form.useWatch("nickname", form);
  const [availability, setAvailability] = useState<{
    nickname: string;
    available: boolean | null;
    checking: boolean;
  }>({ nickname: "", available: null, checking: false });

  const profileQuery = useQuery({
    queryKey: profileQueryKey,
    queryFn: () =>
      graphqlSdk(auth.token)
        .MyProfile()
        .then((data) => data.myProfile),
    enabled: Boolean(auth.token),
    retry: (failureCount, error) =>
      !isSessionInvalidError(error) && failureCount < 3,
  });

  useEffect(() => {
    if (profileQuery.data) {
      form.setFieldsValue({
        nickname: profileQuery.data.nickname ?? "",
        email: auth.user?.email ?? "",
      });
    }
  }, [auth.user?.email, form, profileQuery.data]);

  useEffect(() => {
    if (profileQuery.isError) {
      console.error("Profile loading failed", profileQuery.error);
    }
  }, [profileQuery.error, profileQuery.isError]);

  useEffect(() => {
    const nickname = String(watchedNickname ?? "").trim();
    if (nickname === (profileQuery.data?.nickname ?? "")) {
      setAvailability({ nickname, available: null, checking: false });
      return;
    }
    if (nickname.length < 3 || !auth.token) {
      setAvailability({ nickname, available: null, checking: false });
      return;
    }

    let active = true;
    setAvailability({ nickname, available: null, checking: true });
    const timer = window.setTimeout(() => {
      graphqlSdk(auth.token)
        .NicknameAvailability({ nickname })
        .then((data) => {
          if (
            !active ||
            String(form.getFieldValue("nickname") ?? "").trim() !== nickname
          ) {
            return;
          }
          setAvailability({
            nickname: data.nicknameAvailability.nickname,
            available: data.nicknameAvailability.available,
            checking: false,
          });
        })
        .catch((error) => {
          console.error("Nickname availability check failed", error);
          if (
            !active ||
            String(form.getFieldValue("nickname") ?? "").trim() !== nickname
          ) {
            return;
          }
          setAvailability({ nickname, available: null, checking: false });
        });
    }, 1000);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [auth.token, form, profileQuery.data?.nickname, watchedNickname]);

  const setNicknameMutation = useMutation({
    mutationFn: (nickname: string) =>
      graphqlSdk(auth.token)
        .SetNickname({ nickname })
        .then((data) => data.setNickname),
    onSuccess: (profile) => {
      queryClient.setQueryData(profileQueryKey, profile);
      void queryClient.invalidateQueries({ queryKey: profileQueryKey });
      const savedNickname = profile.nickname ?? "";
      if (auth.user) {
        auth.updateUser({ ...auth.user, nickname: savedNickname });
      }
      setAvailability({
        nickname: savedNickname,
        available: null,
        checking: false,
      });
      void message.success("Профиль сохранён");
    },
    onError: (error) => {
      console.error("Set nickname failed", error);
    },
  });

  const handleSubmit = (values: ProfileFormValues) => {
    if (availability.checking || availability.available === false) {
      return;
    }
    setNicknameMutation.mutate(values.nickname.trim());
  };

  return (
    <section className="max-w-[720px]">
      <Title level={2}>Профиль</Title>

      <Card className="max-w-[480px] rounded-[3px]">
        {profileQuery.isLoading ? (
          <div className="flex justify-center p-12">
            <Spin size="large" />
          </div>
        ) : profileQuery.isError ? (
          <Alert
            message="Не удалось загрузить профиль"
            description="Попробуйте обновить страницу или войти заново."
            type="error"
          />
        ) : (
          <>
            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              requiredMark={false}
            >
              <Form.Item
                label="Nickname"
                name="nickname"
                rules={[
                  { required: true, message: "Укажите nickname" },
                  { min: 2, message: "nickname слишком короткий" },
                ]}
                extra={
                  <NicknameAvailabilityStatus availability={availability} />
                }
              >
                <Input placeholder="aka" />
              </Form.Item>

              <Form.Item label="Email" name="email">
                <Input disabled />
              </Form.Item>

              <Form.Item label="ID пользователя">
                <Text code>{profileQuery.data?.userID ?? auth.user?.id}</Text>
              </Form.Item>

              <Form.Item>
                <Button
                  disabled={
                    availability.checking || availability.available === false
                  }
                  htmlType="submit"
                  type="primary"
                >
                  Сохранить
                </Button>
              </Form.Item>
            </Form>
          </>
        )}
      </Card>
    </section>
  );
}

function NicknameAvailabilityStatus({
  availability,
}: {
  availability: {
    nickname: string;
    available: boolean | null;
    checking: boolean;
  };
}) {
  if (availability.nickname.length < 3) {
    return null;
  }
  if (availability.checking) {
    return (
      <span className="inline-flex items-center gap-2">
        <Spin size="small" />
        <span>Проверяем nickname</span>
      </span>
    );
  }
  if (availability.available === true) {
    return <Text type="success">nickname свободен</Text>;
  }
  if (availability.available === false) {
    return <Text type="danger">nickname занят</Text>;
  }
  return null;
}
