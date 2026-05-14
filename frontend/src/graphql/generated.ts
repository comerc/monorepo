/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
import { GraphQLClient, RequestOptions } from 'graphql-request';
import gql from 'graphql-tag';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
type GraphQLClientRequestHeaders = RequestOptions['requestHeaders'];
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
};

export type LoginPayload = {
  __typename?: 'LoginPayload';
  email: Scalars['String']['output'];
  token: Scalars['String']['output'];
  userID: Scalars['ID']['output'];
};

export type LogoutPayload = {
  __typename?: 'LogoutPayload';
  revoked: Scalars['Boolean']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  loginWithEmailCode: LoginPayload;
  logout: LogoutPayload;
  requestEmailCode: RequestEmailCodePayload;
  setNickname: Profile;
};


export type MutationLoginWithEmailCodeArgs = {
  code: Scalars['String']['input'];
  email: Scalars['String']['input'];
};


export type MutationLogoutArgs = {
  allDevices?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationRequestEmailCodeArgs = {
  email: Scalars['String']['input'];
};


export type MutationSetNicknameArgs = {
  nickname: Scalars['String']['input'];
};

export type NicknameAvailability = {
  __typename?: 'NicknameAvailability';
  available: Scalars['Boolean']['output'];
  nickname: Scalars['String']['output'];
};

export type Profile = {
  __typename?: 'Profile';
  nickname?: Maybe<Scalars['String']['output']>;
  userID: Scalars['ID']['output'];
};

export type Query = {
  __typename?: 'Query';
  authStatus: Scalars['String']['output'];
  myProfile: Profile;
  nicknameAvailability: NicknameAvailability;
  profileStatus: Scalars['String']['output'];
};


export type QueryNicknameAvailabilityArgs = {
  nickname: Scalars['String']['input'];
};

export type RequestEmailCodePayload = {
  __typename?: 'RequestEmailCodePayload';
  accepted: Scalars['Boolean']['output'];
  nextAllowedAt?: Maybe<Scalars['String']['output']>;
  retryAfterSeconds: Scalars['Int']['output'];
};

export type RequestEmailCodeMutationVariables = Exact<{
  email: string;
}>;


export type RequestEmailCodeMutation = { requestEmailCode: { accepted: boolean, retryAfterSeconds: number, nextAllowedAt: string | null } };

export type LoginWithEmailCodeMutationVariables = Exact<{
  email: string;
  code: string;
}>;


export type LoginWithEmailCodeMutation = { loginWithEmailCode: { token: string, userID: string, email: string } };

export type LogoutMutationVariables = Exact<{
  allDevices?: boolean | null | undefined;
}>;


export type LogoutMutation = { logout: { revoked: boolean } };

export type MyProfileQueryVariables = Exact<{ [key: string]: never; }>;


export type MyProfileQuery = { myProfile: { userID: string, nickname: string | null } };

export type NicknameAvailabilityQueryVariables = Exact<{
  nickname: string;
}>;


export type NicknameAvailabilityQuery = { nicknameAvailability: { nickname: string, available: boolean } };

export type SetNicknameMutationVariables = Exact<{
  nickname: string;
}>;


export type SetNicknameMutation = { setNickname: { userID: string, nickname: string | null } };


export const RequestEmailCodeDocument = gql`
    mutation RequestEmailCode($email: String!) {
  requestEmailCode(email: $email) {
    accepted
    retryAfterSeconds
    nextAllowedAt
  }
}
    `;
export const LoginWithEmailCodeDocument = gql`
    mutation LoginWithEmailCode($email: String!, $code: String!) {
  loginWithEmailCode(email: $email, code: $code) {
    token
    userID
    email
  }
}
    `;
export const LogoutDocument = gql`
    mutation Logout($allDevices: Boolean) {
  logout(allDevices: $allDevices) {
    revoked
  }
}
    `;
export const MyProfileDocument = gql`
    query MyProfile {
  myProfile {
    userID
    nickname
  }
}
    `;
export const NicknameAvailabilityDocument = gql`
    query NicknameAvailability($nickname: String!) {
  nicknameAvailability(nickname: $nickname) {
    nickname
    available
  }
}
    `;
export const SetNicknameDocument = gql`
    mutation SetNickname($nickname: String!) {
  setNickname(nickname: $nickname) {
    userID
    nickname
  }
}
    `;

export type SdkFunctionWrapper = <T>(action: (requestHeaders?:Record<string, string>) => Promise<T>, operationName: string, operationType?: string, variables?: any) => Promise<T>;


const defaultWrapper: SdkFunctionWrapper = (action, _operationName, _operationType, _variables) => action();

export function getSdk(client: GraphQLClient, withWrapper: SdkFunctionWrapper = defaultWrapper) {
  return {
    RequestEmailCode(variables: RequestEmailCodeMutationVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<RequestEmailCodeMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<RequestEmailCodeMutation>({ document: RequestEmailCodeDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'RequestEmailCode', 'mutation', variables);
    },
    LoginWithEmailCode(variables: LoginWithEmailCodeMutationVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<LoginWithEmailCodeMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<LoginWithEmailCodeMutation>({ document: LoginWithEmailCodeDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'LoginWithEmailCode', 'mutation', variables);
    },
    Logout(variables?: LogoutMutationVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<LogoutMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<LogoutMutation>({ document: LogoutDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'Logout', 'mutation', variables);
    },
    MyProfile(variables?: MyProfileQueryVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<MyProfileQuery> {
      return withWrapper((wrappedRequestHeaders) => client.request<MyProfileQuery>({ document: MyProfileDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'MyProfile', 'query', variables);
    },
    NicknameAvailability(variables: NicknameAvailabilityQueryVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<NicknameAvailabilityQuery> {
      return withWrapper((wrappedRequestHeaders) => client.request<NicknameAvailabilityQuery>({ document: NicknameAvailabilityDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'NicknameAvailability', 'query', variables);
    },
    SetNickname(variables: SetNicknameMutationVariables, requestHeaders?: GraphQLClientRequestHeaders, signal?: RequestInit['signal']): Promise<SetNicknameMutation> {
      return withWrapper((wrappedRequestHeaders) => client.request<SetNicknameMutation>({ document: SetNicknameDocument, variables, requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders }, signal }), 'SetNickname', 'mutation', variables);
    }
  };
}
export type Sdk = ReturnType<typeof getSdk>;