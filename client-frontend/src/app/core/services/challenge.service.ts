import axios from "axios";
import { ServiceResponse } from "../models/service-response.interface";
import { BaseService } from "./service.base";

export class ChallengeService extends BaseService {
    public async signChallenge(email: string, challenge: string): Promise<ServiceResponse<SignChallengeAsUserResponse>> {
        return this.postAsync<SignChallengeAsUserResponse>('sign', {email: email, challenge: challenge});
    }
}

export interface SignChallengeAsUserResponse {
    signature: string;
}