import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Card, Form, Input, Typography, message } from 'antd'
import { graphqlSdk } from '../graphql/client'
import { useAuth } from '../auth/AuthProvider'
import ThemeSwitcher from '../components/ThemeSwitcher'

const { Text, Title } = Typography

interface LoginFormValues {
  email: string
  code?: string
}

export default function LoginPage() {
  const [form] = Form.useForm<LoginFormValues>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const auth = useAuth()
  const [email, setEmail] = useState('')
  const [codeSent, setCodeSent] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (values: LoginFormValues) => {
    setLoading(true)

    try {
      const normalizedEmail = values.email.trim().toLowerCase()

      if (!codeSent) {
        await graphqlSdk().RequestEmailCode({ email: normalizedEmail })
        setEmail(normalizedEmail)
        setCodeSent(true)
        form.setFieldsValue({ email: normalizedEmail, code: '' })
        void message.success('Код отправлен на email')
        return
      }

      const response = await graphqlSdk().LoginWithEmailCode({
        email: email || normalizedEmail,
        code: values.code?.trim() ?? '',
      })
      const payload = response.loginWithEmailCode

      auth.login(payload.token, {
        id: payload.userID,
        email: payload.email,
        nickname: null,
      })
      navigate(searchParams.get('next') || '/profile')
    } catch (error: unknown) {
      const text = error instanceof Error ? error.message : 'Не удалось войти. Попробуйте снова.'
      message.error(text)
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="auth-page">
      <div className="auth-theme-switcher">
        <ThemeSwitcher />
      </div>
      <Card className="auth-card">
        <div className="auth-title">
          <Title level={3}>Вход</Title>
          <Text type="secondary">
            {codeSent ? `Введите код для ${email}` : 'Получите 4-значный код доступа на email'}
          </Text>
        </div>

        <Form form={form} layout="vertical" onFinish={handleSubmit} autoComplete="off">
          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: 'Укажите email' },
              { type: 'email', message: 'Введите корректный email' },
            ]}
          >
            <Input disabled={codeSent} placeholder="user@example.com" size="large" />
          </Form.Item>

          {codeSent && (
            <Form.Item
              label="Код доступа"
              name="code"
              rules={[
                { required: true, message: 'Введите код доступа' },
                { len: 4, message: 'Код состоит из 4 цифр' },
              ]}
            >
              <Input.OTP length={4} size="large" />
            </Form.Item>
          )}

          <Form.Item>
            <Button block htmlType="submit" loading={loading} size="large" type="primary">
              {codeSent ? 'Войти' : 'Получить код'}
            </Button>
          </Form.Item>
        </Form>

        {codeSent && (
          <Button
            block
            type="link"
            onClick={() => {
              setCodeSent(false)
              form.setFieldsValue({ code: '' })
            }}
          >
            Изменить email
          </Button>
        )}
      </Card>
    </main>
  )
}
