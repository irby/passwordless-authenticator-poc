using AuthenticatorApi.Common.Models.Domain;
using FluentValidation;

namespace AuthenticatorApi.Common.Validators;

public class UserValidator : AbstractValidator<User>
{
    public UserValidator()
    {
        RuleFor(user => user.Username).NotEmpty();
    }
}