using System.Net;
using AuthenticatorApi.Api.Models;
using AuthenticatorApi.Common.Enums;
using AuthenticatorApi.Common.Exceptions;
using FluentValidation;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

namespace AuthenticatorApi.Api.Handlers;

public class ExceptionHandler
{
    private readonly RequestDelegate _next;
    private readonly ILogger<ExceptionHandler> _logger;

    public ExceptionHandler(RequestDelegate next, ILoggerFactory loggerFactory)
    {
        _next = next ?? throw new ArgumentNullException(nameof(next));
        _logger = loggerFactory?.CreateLogger<ExceptionHandler>() ?? throw new ArgumentNullException(nameof(loggerFactory));
    }

    public async Task Invoke(HttpContext context)
    {
        try
        {
            await _next(context);
        }
        catch (ServiceExceptionBase ex)
        {
            if (context.Response.HasStarted)
            {
                throw;
            }

            await LogErrorAndReturnResponse(context, ex, ex.ErrorCode, ex.ResponseCode, ex.DefaultMessage);
        }
        catch (ValidationException ex)
        {
            await LogErrorAndReturnResponse(context, ex, null, (int)HttpStatusCode.BadRequest, string.Join(", ", ex.Errors.Select(p => p.ErrorMessage)));
        }
        catch (Exception ex)
        {
            await LogErrorAndReturnResponse(context, ex, null, (int)HttpStatusCode.InternalServerError);
        }
    }

    private async Task LogErrorAndReturnResponse(HttpContext context, Exception ex, ErrorCode? errorCode, int statusCode, string? errorMessage = null)
    {
        _logger.LogError(ex, ex.Message);
            
        context.Response.Clear();
        context.Response.StatusCode = statusCode;
        context.Response.ContentType = "application/json";
        await context.Response.WriteAsync(JsonConvert.SerializeObject(new ErrorResponse
        {
            ErrorCode = errorCode,
            ErrorMessage = errorMessage ?? ex.Message,
            StatusCode = statusCode
        }, new JsonSerializerSettings { ContractResolver = new CamelCasePropertyNamesContractResolver() }));
    }
}

public static class ExceptionHandlerExtensions
{
    public static IApplicationBuilder UseServiceExceptionHandler(this IApplicationBuilder builder)
    {
        return builder.UseMiddleware<ExceptionHandler>();
    }
}