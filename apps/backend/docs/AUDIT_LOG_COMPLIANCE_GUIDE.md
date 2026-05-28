# Audit Log Compliance Guide (Badan POM)

## Overview

This guide provides comprehensive information about audit log compliance requirements for Badan POM (National Agency of Drug and Food Control) regulatory inspections in Indonesia.

**Document Version:** 1.0  
**Last Updated:** 2026-05-27  
**Applies to:** Simpo Pharmacy Management System v1.0+

---

## Regulatory Requirements

### Badan POM Audit Trail Standards

According to Badan POM regulations, pharmacy management systems must maintain:

1. **Append-Only Audit Trail**: All system changes must be logged with no modification or deletion capabilities
2. **User Identification**: Every audit entry must identify who performed the action
3. **Timestamp**: All entries must include accurate date and time information
4. **Reason Recording**: Every action must include the reason for the change
5. **Minimum 5-Year Retention**: Audit logs must be retained for at least 5 years per regulatory requirements

### Compliance Standards

- **NFR-SEC-004**: Append-only audit trail for all system changes with user identification, timestamp, and reason
- **NFR-SEC-009**: Complete audit trail for all inventory transactions for minimum 5 years
- **NFR-SEC-010**: Read-only audit mode for compliance verification without data alteration capabilities

---

## Audit Log Categories

### 1. Authentication & Authorization
- `LOGIN_SUCCESS` - Successful user login
- `LOGIN_FAILURE` - Failed login attempts
- `LOGOUT` - User logout
- `PASSWORD_RESET` - Password reset operations
- `AUTH_FAILURE` - Authentication failures
- `FORBIDDEN_ACCESS` - Unauthorized access attempts

### 2. User Management
- `USER_CREATED` - New user account creation
- `USER_DEACTIVATED` - User account deactivation
- `SELF_REGISTRATION` - Staff self-registration
- `EMAIL_VERIFIED` - Email verification completion
- `ROLE_UPDATED` - Role changes (Story 6.4)
- `PERMISSION_GRANTED` - Permission grants (Story 6.4)
- `PERMISSION_REVOKED` - Permission revocations (Story 6.4)

### 3. Inventory Management
- `STOCK_ADJUSTMENT` - Manual stock adjustments
- `BLOCKED_SALE_ATTEMPT` - Blocked sales of expired products

### 4. System Configuration
- `SYSTEM_SETTINGS_UPDATED` - Pharmacy settings changes (Story 6.4)
- `WHITELIST_DOMAIN_ADDED` - Whitelist domain additions
- `WHITELIST_DOMAIN_UPDATED` - Whitelist domain updates
- `WHITELIST_DOMAIN_DELETED` - Whitelist domain deletions

### 5. Backup & Recovery
- `BACKUP_CREATED` - Backup creation operations (Story 6.4)
- `BACKUP_RESTORED` - Backup restore operations (Story 6.4)
- `BACKUP_DELETED` - Backup deletion operations (Story 6.4)

### 6. Branch Management
- `BRANCH_CREATED` - New branch creation (Story 6.4)
- `BRANCH_UPDATED` - Branch information updates (Story 6.4)
- `BRANCH_DEACTIVATED` - Branch deactivation (Story 6.4)

### 7. System Operations
- `SYSTEM_STARTUP` - Application startup (Story 6.4)
- `SYSTEM_SHUTDOWN` - Application shutdown (Story 6.4)
- `MAINTENANCE_MODE_ENABLED` - Maintenance mode activation (Story 6.4)
- `MAINTENANCE_MODE_DISABLED` - Maintenance mode deactivation (Story 6.4)

### 8. Reporting
- `EXPORT_REPORT` - Report generation and export

---

## Data Retention Policy

### 5-Year Minimum Retention

**Regulatory Requirement**: All audit logs must be retained for a minimum of 5 years from the date of creation.

**Implementation Details**:
- Database: PostgreSQL with time-based partitioning
- Automatic Cleanup: Scheduled job runs weekly to identify records older than 5 years
- Cleanup Execution: Only SystemAdmin role can execute cleanup with explicit confirmation
- Backup Before Cleanup: All audit log deletions are backed up before removal

**Retention Calculation**:
```
Retention Date = Current Date - 5 Years
Example: 2026-05-27 - 5 years = 2021-05-27
Records from 2021-05-26 and earlier are eligible for cleanup
```

**Boundary Case Handling**:
- Records exactly 5 years old are retained (not deleted)
- Only records older than 5 years are removed
- Example: On 2026-05-27, records from 2021-05-27 are kept, records from 2021-05-26 are deleted

**Audit Cleanup API**:
```http
POST /api/v1/audit/cleanup?confirm=true
Authorization: Bearer <system_admin_token>
```

---

## Audit Log Data Structure

### Standard Fields

Every audit log entry contains:
- `id` - Unique identifier (auto-increment)
- `timestamp` - Date and time of action (ISO 8601 format)
- `user_id` - ID of user who performed action
- `username` - Username of performer
- `action` - Type of action performed (see categories above)
- `ip_address` - IP address of performer
- `outcome` - Result of action (success/failure/blocked/pending)
- `reason` - Detailed reason for action

### System Change Specific Fields

For system configuration changes (Story 6.4):
- `old_value` - Previous value before change
- `new_value` - New value after change
- `change_details` - JSON structure with complete change information

---

## Access Control & Security

### RBAC Requirements

**Roles with Audit Log Access**:
- ✅ **SYSTEM_ADMIN** - Full access to all audit logs
- ✅ **OWNER** - Full access to all audit logs
- ✅ **ADMIN** - Full access to all audit logs
- ❌ **CASHIER** - NO access (business-sensitive data)

**Query Permissions**:
- Required authentication via JWT token
- Date range filters mandatory (max 1 year per query)
- IP address logging enforced for all audit entries

### Append-Only Enforcement

**Database Level**:
- INSERT-only permissions on audit_logs table
- No UPDATE or DELETE permissions at database level
- Application-level validation rejects modification attempts

**Application Level**:
- AuditRepository interface has no Update/Delete methods
- Only Create and Query methods available
- Runtime validation prevents any modification attempts

---

## Compliance Inspection Preparation

### Pre-Inspection Checklist

- [ ] Verify audit log retention (5 years minimum)
- [ ] Test audit log export functionality (CSV/PDF)
- [ ] Confirm append-only behavior (no modification possible)
- [ ] Validate user identification in all entries
- [ ] Check timestamp accuracy (timezone: Asia/Jakarta)
- [ ] Ensure reason field is populated for all manual changes
- [ ] Verify RBAC controls (cashier access denied)
- [ ] Test audit log query interface with various filters

### Required Documentation for Inspections

1. **Audit Log Retention Policy** - This document
2. **System Architecture Documentation** - Database schema and append-only design
3. **Access Control Matrix** - RBAC permissions and user roles
4. **Audit Log Export Procedure** - Step-by-step guide for data extraction
5. **Sample Audit Logs** - Examples of each audit action type

### Common Inspection Scenarios

**Scenario 1: Verify Append-Only Behavior**
```
Inspector Request: "Show that audit logs cannot be modified"
Response: 
- Demonstrate repository interface (no Update/Delete methods)
- Show database permissions (INSERT-only)
- Attempt modification via API (should fail)
- Provide audit log export as evidence of integrity
```

**Scenario 2: Audit Trail for Specific Product**
```
Inspector Request: "Show all changes for product PARACETAMOL in last 6 months"
Response:
- Query audit logs by date range and product SKU
- Export results to CSV/PDF
- Include all stock adjustments, blocked sales, and configuration changes
```

**Scenario 3: User Activity Timeline**
```
Inspector Request: "Show all actions performed by user 'admin' on 2026-05-27"
Response:
- Filter by username and specific date
- Include IP address, actions, and reasons
- Export to CSV with complete timeline
```

---

## Best Practices

### For System Administrators

1. **Regular Backups**: Backup audit logs before any retention cleanup
2. **Monitor Storage**: Ensure sufficient database storage for 5+ years of logs
3. **Access Reviews**: Quarterly review of audit log access permissions
4. **Performance Tuning**: Monitor query performance for large date ranges
5. **Compliance Audits**: Annual internal audit of audit log integrity

### For Pharmacy Owners

1. **Understand Requirements**: Know Badan POM audit trail requirements
2. **Staff Training**: Train staff on audit log importance and compliance
3. **Regular Reviews**: Review audit logs monthly for suspicious activity
4. **Incident Response**: Have procedures for investigating compliance issues
5. **Documentation**: Maintain records of compliance activities

---

## Troubleshooting

### Common Issues

**Issue**: Audit logs not appearing in query results
**Solution**: 
- Verify date range is correct (max 1 year)
- Check user has appropriate RBAC permissions
- Ensure action filter is not too restrictive

**Issue**: Export failing for large date ranges
**Solution**:
- Reduce date range to 1 year or less
- Use action filters to reduce result set
- Increase server timeout if needed

**Issue**: Cannot access audit logs page
**Solution**:
- Verify user role (Admin, Owner, SystemAdmin only)
- Check JWT token is valid and not expired
- Ensure user is not in Cashier role

---

## Contact & Support

**Technical Support**: support@simpopharmacy.com  
**Compliance Questions**: compliance@simpopharmacy.com  
**Documentation**: https://docs.simpopharmacy.com/audit-logs

---

## Revision History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-05-27 | Initial release - Badan POM compliance guide | Development Team |

---

**Document Classification**: Internal - Confidential  
**Distribution**: System Administrators, Pharmacy Owners, Compliance Officers
