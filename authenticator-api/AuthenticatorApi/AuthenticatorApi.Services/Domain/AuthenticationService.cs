using AuthenticatorApi.Common.Enums;
using AuthenticatorApi.Common.Exceptions;
using AuthenticatorApi.Common.Models.Dto;
using AuthenticatorApi.Common.Models.Dto.Authentication;
using AuthenticatorApi.Common.Validators.User;
using FluentValidation;

namespace AuthenticatorApi.Services.Domain;

public class AuthenticationService : ServiceBase<AuthenticationService>
{
    private IUserService _userService;

    public AuthenticationService(ILoggerFactory loggerFactory, IUserService userService, IMapper mapper) : base(loggerFactory, mapper)
    {
        _userService = userService;
    }

    public async Task LoginUserAsync(UserLoginDto dto)
    {
        await ValidateBaseRequest(dto);
        
        var trackingId = Guid.NewGuid();
        
        Log.LogInformation($"Starting login for user {dto.Username} at tenant ID {dto.TenantId}. Tracking ID: {trackingId}");
        
        var user = await _userService.GetUserByTenantIdAndUsername(dto.TenantId, dto.Username);

        if (user is null)
        {
            Log.LogInformation($"Login failed for {dto.Username}. Reason: Does not exist. Tracking ID: {trackingId}");
            throw new NotFoundException();
        }

        if (!user.IsActive)
        {
            Log.LogInformation($"Login failed for {dto.Username}. Reason: User is not active. Tracking ID: {trackingId}");
            throw new BadRequestException(ErrorCode.AccountDisabled);
        }

        if (!user.IsVerified)
        {
            Log.LogInformation($"Login failed for {dto.Username}. Reason: User is not yet verified. Tracking ID: {trackingId}");
            throw new BadRequestException(ErrorCode.AccountNotVerified);
        }
        
        Log.LogInformation($"Login successful for {dto.Username}. Tracking ID: {trackingId}");
    }

    public async Task<CreateUserResponseDto> RegisterUser(CreateUserRequestDto dto)
    {
        await ValidateBaseRequest(dto);
        
        var existingUser = await _userService.GetUserByTenantIdAndUsername(dto.TenantId, dto.Username);

        if (existingUser is { })
        {
            throw new BadRequestException(ErrorCode.AccountAlreadyExists);
        }

        var user = await _userService.CreateUser(dto);

        return new CreateUserResponseDto()
        {
            Id = user.Id,
            CreatedOn = user.CreatedOn.GetValueOrDefault(),
            ModifiedOn = user.ModifiedOn.GetValueOrDefault(),
            IsVerified = user.IsVerified
        };
    }

    private async Task ValidateBaseRequest(BaseUserRequestDto dto)
    {
        await new UserRequestDtoValidator().ValidateAndThrowAsync(dto);
    }
}