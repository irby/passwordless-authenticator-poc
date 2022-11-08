import { ErrorCode } from "../enums/error-code.enum";

export interface ErrorResponse {
  statusCode: number;
  errorMessage: string;
  errorCode: ErrorCode;
}
