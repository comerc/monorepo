import { GraphQLClient } from 'graphql-request'
import { getSdk } from './generated'

const endpoint =
  import.meta.env.VITE_GRAPHQL_URL || new URL('/graphql', window.location.origin).toString()

export function graphqlSdk(token?: string | null) {
  const client = new GraphQLClient(endpoint, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  })
  return getSdk(client)
}
