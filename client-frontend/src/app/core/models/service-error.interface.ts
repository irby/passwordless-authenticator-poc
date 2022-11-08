import { ErrorResponse } from "./error-response.interface";

export interface ServiceError {
  type: "error";
  statusCode: number;
  message: string;
  body: string;
  response: ErrorResponse;
}
