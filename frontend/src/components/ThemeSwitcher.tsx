import { DesktopOutlined, MoonOutlined, SunOutlined } from '@ant-design/icons'
import { Segmented, Tooltip } from 'antd'
import { useThemeMode, type ThemeMode } from '../theme/ThemeProvider'

export default function ThemeSwitcher() {
  const { mode, setMode } = useThemeMode()

  return (
    <Segmented<ThemeMode>
      value={mode}
      onChange={setMode}
      size="small"
      options={[
        {
          value: 'light',
          icon: (
            <Tooltip title="Светлая тема">
              <SunOutlined />
            </Tooltip>
          ),
        },
        {
          value: 'auto',
          icon: (
            <Tooltip title="Как в системе">
              <DesktopOutlined />
            </Tooltip>
          ),
        },
        {
          value: 'dark',
          icon: (
            <Tooltip title="Тёмная тема">
              <MoonOutlined />
            </Tooltip>
          ),
        },
      ]}
    />
  )
}
