import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: [
    '../backend/auth/internal/transport/http/graphql/schema.graphqls',
    '../backend/profile/internal/transport/http/graphql/schema.graphqls',
  ],
  documents: ['src/graphql/**/*.graphql'],
  generates: {
    'src/graphql/generated.ts': {
      plugins: ['typescript', 'typescript-operations', 'typescript-graphql-request'],
    },
  },
}

export default config
