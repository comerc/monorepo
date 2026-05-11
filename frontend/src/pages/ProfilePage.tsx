import { useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Alert, Button, Card, Form, Input, Spin, Typography, message } from 'antd'
import { useAuth } from '../auth/AuthProvider'
import { graphqlSdk } from '../graphql/client'

const { Text, Title } = Typography

interface ProfileFormValues {
  nickname: string
  email: string
}

export default function ProfilePage() {
  const [form] = Form.useForm<ProfileFormValues>()
  const queryClient = useQueryClient()
  const auth = useAuth()

  const profileQuery = useQuery({
    queryKey: ['my-profile', auth.user?.id],
    queryFn: () => graphqlSdk(auth.token).MyProfile().then((data) => data.myProfile),
    enabled: Boolean(auth.token),
  })

  useEffect(() => {
    if (profileQuery.data) {
      form.setFieldsValue({
        nickname: profileQuery.data.nickname ?? '',
        email: auth.user?.email ?? '',
      })
    }
  }, [auth.user?.email, form, profileQuery.data])

  const setNicknameMutation = useMutation({
    mutationFn: (nickname: string) =>
      graphqlSdk(auth.token).SetNickname({ nickname }).then((data) => data.setNickname),
    onSuccess: (profile) => {
      void queryClient.invalidateQueries({ queryKey: ['my-profile', auth.user?.id] })
      if (auth.user) {
        auth.updateUser({ ...auth.user, nickname: profile.nickname })
      }
      void message.success('Профиль сохранён')
    },
  })

  const handleSubmit = (values: ProfileFormValues) => {
    setNicknameMutation.mutate(values.nickname.trim())
  }

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
            description={(profileQuery.error as Error).message}
            type="error"
          />
        ) : (
          <>
            {setNicknameMutation.isError && (
              <Alert
                closable
                className="mb-4"
                message="Не удалось сохранить профиль"
                description={(setNicknameMutation.error as Error).message}
                type="error"
              />
            )}

            <Form form={form} layout="vertical" onFinish={handleSubmit}>
              <Form.Item
                label="Nickname"
                name="nickname"
                rules={[
                  { required: true, message: 'Укажите nickname' },
                  { min: 2, message: 'Nickname должен быть не короче 2 символов' },
                ]}
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
                  htmlType="submit"
                  loading={setNicknameMutation.isPending}
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
  )
}
