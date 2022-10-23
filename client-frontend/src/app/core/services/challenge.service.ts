import axios from "axios";
import { ServiceResponse } from "../models/service-response.interface";

export class ChallengeService {
    public async signChallenge(email: string, challenge: string) {
        return await axios.post('http://localhost:3000', {email: email, challenge: challenge});
    }
}