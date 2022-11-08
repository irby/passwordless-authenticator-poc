export interface WebAuthnLoginFinalizeRequest {
  type: string;
  id: string;
  rawId: string;
  authenticationAttachment: string;
  response: WebAuthnLoginFinalizeRequestResponse;
  clientExtensionResults: any;
}

export interface WebAuthnLoginFinalizeRequestResponse {
  clientDataJSON: string;
  authenticatorData: string;
  signature: string;
  userHandle: string;
}

export function GenerateWebAuthnLoginFinalizeRequest(): WebAuthnLoginFinalizeRequest {
  const request: WebAuthnLoginFinalizeRequest = {
    type: "public-key",
    id: "",
    rawId: "",
    authenticationAttachment: "platform",
    response: {
      clientDataJSON: "",
      authenticatorData: "",
      signature: "",
      userHandle: "",
    },
    clientExtensionResults: {},
  };
  return request;
}
