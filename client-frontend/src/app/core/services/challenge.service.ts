import { Injectable } from "@angular/core";
import axios from "axios";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

@Injectable()
export class ChallengeService extends BaseService {
  public async signChallenge(
    userId: string,
    challenge: string
  ): Promise<ServiceResponse<SignChallengeAsUserResponse>> {
    return this.postAsync<SignChallengeAsUserResponse>("sign", {
      userId: userId,
      challenge: challenge,
    });
  }
}

export interface SignChallengeAsUserResponse {
  id: string;
  signature: string;
  clientDataJson: string;
  authenticatorData: string;
  userHandle: string;
}
