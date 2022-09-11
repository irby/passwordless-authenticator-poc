using Microsoft.AspNetCore.Mvc;

namespace AuthenticatorApi.Api.Controllers;

[ResponseCache(NoStore = true, Location = ResponseCacheLocation.None, Duration = -1)]
[Route("api/[controller]")]
public abstract class ApiControllerBase : ControllerBase
{
    protected new IActionResult Ok()
    {
        return Ok(new EmptyResult());
    }
}