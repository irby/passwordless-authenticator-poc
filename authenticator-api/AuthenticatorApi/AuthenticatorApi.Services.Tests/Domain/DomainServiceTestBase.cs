using AuthenticatorApi.Common.Interfaces;
using AuthenticatorApi.Data;
using AuthenticatorApi.Services.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace AuthenticatorApi.Services.Tests.Domain;

public abstract class DomainServiceTestBase<T> : ServiceTestBase<T> where T : DomainServiceBase<T>
{
    public Guid? CurrentUserId { get; set; }

    protected DomainServiceTestBase()
    {
        ApplicationUserResolverMock.Setup(p => p.TryGetUserId()).Returns(CurrentUserId);
    }
}