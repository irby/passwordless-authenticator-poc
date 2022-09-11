using AuthenticatorApi.Common.Enums;

namespace AuthenticatorApi.Api.Models;

public class ErrorResponse
{
    public int StatusCode { get; set; }
    public ErrorCode? ErrorCode { get; set; }
    public string ErrorMessage { get; set; }
}