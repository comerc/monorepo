import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { App as AntdApp, Button, Card, Form, Input, Typography } from "antd";
import { graphqlSdk } from "../graphql/client";
import { useAuth } from "../auth/authStore";
import ThemeSwitcher from "../components/ThemeSwitcher";

const { Text, Title } = Typography;

interface LoginFormValues {
  email: string;
  code?: string;
}

export default function LoginPage() {
  const [form] = Form.useForm<LoginFormValues>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const auth = useAuth();
  const [email, setEmail] = useState("");
  const [codeSent, setCodeSent] = useState(false);
  const [requestLoading, setRequestLoading] = useState(false);
  const [loginLoading, setLoginLoading] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);
  const [retryAfterSeconds, setRetryAfterSeconds] = useState(0);
  const { message } = AntdApp.useApp();

  useEffect(() => {
    if (retryAfterSeconds <= 0) {
      return;
    }
    const timer = window.setInterval(() => {
      setRetryAfterSeconds((seconds) => Math.max(0, seconds - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [retryAfterSeconds]);

  const requestCode = async (requestedEmail: string) => {
    const response = await graphqlSdk().RequestEmailCode({
      email: requestedEmail,
    });
    const payload = response.requestEmailCode;
    setRetryAfterSeconds(payload.retryAfterSeconds);
    if (!payload.accepted) {
      void message.error(
        payload.retryAfterSeconds > 0
          ? `Следующий код можно запросить через ${payload.retryAfterSeconds} сек.`
          : "Не удалось запросить код. Попробуйте снова.",
      );
      return false;
    }
    return true;
  };

  const loginWithCode = async (code: string) => {
    if (loginLoading || code.length !== 5) {
      return;
    }
    setLoginLoading(true);
    try {
      const response = await graphqlSdk().LoginWithEmailCode({
        email,
        code,
      });
      const payload = response.loginWithEmailCode;

      auth.login(payload.token, {
        id: payload.userID,
        email: payload.email,
        nickname: null,
      });
      navigate(searchParams.get("next") || "/profile");
    } catch (error: unknown) {
      const text =
        error instanceof Error
          ? error.message
          : "Не удалось войти. Попробуйте снова.";
      void message.error(text);
    } finally {
      setLoginLoading(false);
    }
  };

  const handleSubmit = async (values: LoginFormValues) => {
    if (codeSent) {
      return;
    }
    setRequestLoading(true);

    try {
      const normalizedEmail = values.email.trim().toLowerCase();
      const accepted = await requestCode(normalizedEmail);
      if (!accepted) {
        return;
      }
      setEmail(normalizedEmail);
      setCodeSent(true);
      form.setFieldsValue({ code: "" });
      void message.success("Код отправлен на email");
    } catch (error: unknown) {
      const text =
        error instanceof Error
          ? error.message
          : "Не удалось запросить код. Попробуйте снова.";
      void message.error(text);
    } finally {
      setRequestLoading(false);
    }
  };

  const handleResend = async () => {
    if (retryAfterSeconds > 0) {
      return;
    }
    setResendLoading(true);
    try {
      await requestCode(email);
    } catch (error: unknown) {
      const text =
        error instanceof Error
          ? error.message
          : "Не удалось запросить код. Попробуйте снова.";
      void message.error(text);
    } finally {
      setResendLoading(false);
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-app-bg p-6 dark:bg-app-bg-dark">
      <div className="fixed right-4 top-4">
        <ThemeSwitcher />
      </div>
      <Card className="w-full max-w-[400px] rounded-[3px]">
        <div className="mb-6 text-center">
          <Title level={3}>Вход</Title>
          <Text type="secondary">
            {codeSent
              ? `Введите код для ${email}`
              : "Получите 5-значный код доступа на email"}
          </Text>
        </div>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          requiredMark={false}
        >
          {!codeSent && (
            <Form.Item
              label="Email"
              name="email"
              rules={[
                { required: true, message: "Введите email" },
                { type: "email", message: "email is invalid" },
              ]}
              validateTrigger={[]}
            >
              <Input placeholder="user@example.com" size="large" />
            </Form.Item>
          )}

          {codeSent && (
            <Form.Item
              label="Код доступа"
              name="code"
              rules={[
                { required: true, message: "Введите код доступа" },
                { len: 5, message: "Код состоит из 5 цифр" },
              ]}
            >
              <Input.OTP
                disabled={loginLoading}
                length={5}
                size="large"
                onChange={(value) => {
                  void loginWithCode(value);
                }}
              />
            </Form.Item>
          )}

          {!codeSent && (
            <Form.Item>
              <Button
                block
                htmlType="submit"
                loading={requestLoading}
                size="large"
                type="primary"
              >
                Получить код
              </Button>
            </Form.Item>
          )}
        </Form>

        {codeSent && (
          <div className="flex flex-col items-start gap-2">
            {retryAfterSeconds > 0 && (
              <div>
                <Text type="secondary">
                  Следующий код можно запросить через {retryAfterSeconds} сек.
                </Text>
              </div>
            )}
            {retryAfterSeconds <= 0 && (
              <Button
                className="px-0"
                loading={resendLoading}
                type="link"
                onClick={handleResend}
              >
                Отправить код ещё раз
              </Button>
            )}
            <Button
              className="px-0"
              type="link"
              onClick={() => {
                setCodeSent(false);
                setRetryAfterSeconds(0);
                form.setFieldsValue({ code: "" });
              }}
            >
              Изменить email
            </Button>
          </div>
        )}
      </Card>
    </main>
  );
}
