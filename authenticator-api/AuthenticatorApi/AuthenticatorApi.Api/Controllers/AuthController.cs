using AuthenticatorApi.Common.Models.Dto.Authentication;
using AuthenticatorApi.Services.Domain;
using Microsoft.AspNetCore.Mvc;

namespace AuthenticatorApi.Api.Controllers;

public class AuthController : ApiControllerBase
{
    private readonly AuthenticationService _authenticationService;
    
    public AuthController(AuthenticationService authenticationService)
    {
        _authenticationService = authenticationService;
    }
    
    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] UserLoginDto dto)
    {
        var result = await _authenticationService.LoginUserAsync(dto.TenantId, dto.Username);
        return Ok(result);
    }
}