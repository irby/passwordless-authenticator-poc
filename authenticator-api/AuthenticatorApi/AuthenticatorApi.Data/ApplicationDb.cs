using AuthenticatorApi.Common.Constants;
using AuthenticatorApi.Common.Interfaces;
using AuthenticatorApi.Common.Models;
using AuthenticatorApi.Common.Models.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Diagnostics;

namespace AuthenticatorApi.Data;

public class ApplicationDb : DbContext
{
    private static readonly EntityState[] AuditableStates =
    {
        EntityState.Added, EntityState.Deleted, EntityState.Modified
    };
    private readonly IApplicationUserResolver _applicationUserResolver;

    public ApplicationDb(DbContextOptions<ApplicationDb> options, IApplicationUserResolver applicationUserResolver) : base(options)
    {
        _applicationUserResolver = applicationUserResolver;
    }
    
    public virtual DbSet<User> Users { get; set; }
    public virtual DbSet<UserPublicKey> UserPublicKeys { get; set; }
    public virtual DbSet<Tenant> Tenants { get; set; }
    public virtual DbSet<TenantAllowedDomain> TenantAllowedDomains { get; set; }

    protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
    {
        if (optionsBuilder.Options.Extensions.FirstOrDefault(p => p.Info.LogFragment.Contains(StringConstants.AuthenticationDbName)) != null)
        {
            optionsBuilder.UseInMemoryDatabase("name=DB");
        }
        
        // Ignore Include warnings on EF navigation
        optionsBuilder.ConfigureWarnings(opt => opt.Ignore(CoreEventId.NavigationBaseIncludeIgnored));
    }

    public new async Task<int> SaveChangesAsync(bool skipAuditing = false, CancellationToken cancellationToken = new CancellationToken())
    {
        var auditableEntities = new List<Tuple<AuditableEntity, EntityState>>();
        if (!skipAuditing)
        {
            auditableEntities = PrepareAuditableEntities();
        }
        var retVal = await base.SaveChangesAsync(cancellationToken);
        if (!skipAuditing)
        {
            await UpdateAuditableEntitiesAsync(auditableEntities);
        }
        
        return retVal;
    }
    
    /// <summary>
    /// Fetch the auditable entries and set their modified / created values based on their state
    /// </summary>
    /// <returns></returns>
    private List<Tuple<AuditableEntity, EntityState>> PrepareAuditableEntities()
    {
        var auditableEntries = ChangeTracker.Entries()
            .Where(x => x.Entity is AuditableEntity && AuditableStates.Contains(x.State))
            .Select(x => new Tuple<AuditableEntity?, EntityState>(x.Entity as AuditableEntity, x.State)).ToList();

        foreach (var entry in auditableEntries)
        {
            if (entry.Item2 == EntityState.Modified)
            {
                entry.Item1.ModifiedOn = DateTimeOffset.UtcNow;
                entry.Item1.ModifiedBy = _applicationUserResolver.TryGetUserId();
            }
            else if (entry.Item2 == EntityState.Added)
            {
                entry.Item1.ModifiedOn = entry.Item1.CreatedOn = DateTimeOffset.UtcNow;
                entry.Item1.ModifiedBy = entry.Item1.CreatedBy = _applicationUserResolver.TryGetUserId();
            }
        }

        return auditableEntries;
    }

    private async Task UpdateAuditableEntitiesAsync(List<Tuple<AuditableEntity, EntityState>> entries)
    {
        foreach (var (auditableEntity, entityState) in entries)
        {
            // TODO: complete this
            // await _auditRecorder.RecordAuditEntryAsync(auditableEntity, entityState,
            //     _applicationUserResolver.TryGetUserId());
        }
    }
}