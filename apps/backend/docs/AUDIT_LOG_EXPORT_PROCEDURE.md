# Audit Log Export Procedure for Badan POM Inspections

## Overview

This document provides step-by-step procedures for exporting audit logs during Badan POM regulatory inspections.

**Document Version:** 1.0  
**Last Updated:** 2026-05-27  
**Target Audience**: System Administrators, Pharmacy Owners, Compliance Officers

---

## Export Methods

### Method 1: Web Dashboard Export (Recommended)

**Best For**: Quick exports, specific date ranges, ad-hoc inspections

**Steps**:
1. Log in to Simpo Pharmacy Management System
2. Navigate to **Admin** → **Audit Logs**
3. Set date range filters (maximum 1 year per export)
4. Optionally filter by action type or category
5. Click **Export CSV** button
6. Save file with descriptive name: `AuditLogs_[START_DATE]_to_[END_DATE].csv`

**Example URL**:
```
https://simpopharmacy.com/admin/audit-logs
```

### Method 2: API Export

**Best For**: Automated exports, large date ranges, integration with other systems

**Endpoint**:
```http
GET /api/v1/audit/logs/export?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&format=csv
Authorization: Bearer <your_jwt_token>
```

**cURL Example**:
```bash
curl -X GET "https://api.simpopharmacy.com/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=csv" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Accept: text/csv" \
  --output audit_logs_2026.csv
```

**Python Example**:
```python
import requests

headers = {
    'Authorization': f'Bearer {jwt_token}',
    'Accept': 'text/csv'
}

params = {
    'start_date': '2026-01-01',
    'end_date': '2026-12-31',
    'format': 'csv'
}

response = requests.get(
    'https://api.simpopharmacy.com/api/v1/audit/logs/export',
    headers=headers,
    params=params
)

with open('audit_logs_2026.csv', 'wb') as f:
    f.write(response.content)
```

---

## Export Formats

### CSV Format (Comma-Separated Values)

**File Extension**: `.csv`  
**Content Type**: `text/csv`  
**Best For**: Excel analysis, data import, further processing

**Column Structure**:
```csv
id,timestamp,user_id,username,action,ip_address,outcome,reason
1,2026-05-27T10:30:00+07:00,1,admin,SYSTEM_SETTINGS_UPDATED,192.168.1.100,success,"{""pharmacy_name"":""Simpo Pharmacy""}"
2,2026-05-27T11:15:00+07:00,1,admin,BACKUP_CREATED,192.168.1.100,success,"simpo_20260527_111500.dump (1024000 bytes)"
3,2026-05-27T12:00:00+07:00,2,pharmacist,STOCK_ADJUSTMENT,192.168.1.101,success,"PARACETAMOL: 100 → 95 (Damaged packaging)"
```

### JSON Format

**File Extension**: `.json`  
**Content Type**: `application/json`  
**Best For**: Programmatic processing, API integration

**Structure**:
```json
[
  {
    "id": 1,
    "timestamp": "2026-05-27T10:30:00+07:00",
    "user_id": 1,
    "username": "admin",
    "action": "SYSTEM_SETTINGS_UPDATED",
    "ip_address": "192.168.1.100",
    "outcome": "success",
    "reason": "{\"pharmacy_name\":\"Simpo Pharmacy\"}"
  }
]
```

---

## Pre-Export Preparation

### Step 1: Determine Date Range

**Badan POM Inspection Requirements**:
- Typical inspection period: Last 12-24 months
- Full compliance audit: Up to 5 years (may require multiple exports)
- Specific incident: Exact date range of the incident

**Date Range Constraints**:
- Maximum range per export: 1 year (365 days)
- For longer periods, perform multiple exports
- Example: 3-year audit = 3 separate exports (Year 1, Year 2, Year 3)

### Step 2: Verify Access Permissions

**Required Role**: One of the following
- ✅ SYSTEM_ADMIN
- ✅ OWNER  
- ✅ ADMIN
- ❌ CASHIER (access denied)

**Verification**:
```bash
# Check current user role
GET /api/v1/auth/me
Authorization: Bearer <your_jwt_token>

# Expected response includes "role": "ADMIN" or higher
```

### Step 3: Plan Export Strategy

**For 1-Year Inspection**:
- Single export: `start_date=2025-05-27&end_date=2026-05-27`

**For 3-Year Compliance Audit**:
- Export 1: `start_date=2023-05-27&end_date=2024-05-27`
- Export 2: `start_date=2024-05-27&end_date=2025-05-27`
- Export 3: `start_date=2025-05-27&end_date=2026-05-27`

**For Specific Investigation**:
- Use exact incident dates
- Include 24 hours before and after for context

---

## Export Scenarios

### Scenario 1: Annual Compliance Export

**Purpose**: Prepare annual audit log package for compliance records  
**Frequency**: Once per year  
**Date Range**: Full calendar year (Jan 1 to Dec 31)

**Procedure**:
```bash
# Export full year 2026
curl -X GET "https://api.simpopharmacy.com/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=csv" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -o "AuditLogs_2026_FullYear.csv"

# Verify file size (should be substantial for full year)
ls -lh AuditLogs_2026_FullYear.csv

# Calculate checksum for integrity verification
shasum AuditLogs_2026_FullYear.csv > AuditLogs_2026_FullYear.csv.sha256
```

**File Naming Convention**:
```
AuditLogs_YYYY_FullYear.csv
Example: AuditLogs_2026_FullYear.csv
```

### Scenario 2: Incident Investigation Export

**Purpose**: Extract audit logs for specific incident or investigation  
**Trigger**: Ad-hoc request from Badan POM or internal investigation  
**Date Range**: 24 hours before to 24 hours after incident

**Procedure**:
```bash
# Example: Stock discrepancy incident on 2026-05-15 at 14:30
INCIDENT_DATE="2026-05-15"

# Export 24-hour window around incident
curl -X GET "https://api.simpopharmacy.com/api/v1/audit/logs/export?start_date=2026-05-14&end_date=2026-05-16&format=csv" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -o "Incident_${INCIDENT_DATE}_AuditTrail.csv"

# Filter for relevant actions (stock adjustments, user activity)
# Use spreadsheet software to filter by 'action' column
```

**File Naming Convention**:
```
Incident_YYYY-MM-DD_AuditTrail.csv
Example: Incident_2026-05-15_AuditTrail.csv
```

### Scenario 3: User Activity Export

**Purpose**: Provide timeline of specific user's actions  
**Use Case**: Employee investigation, performance review, suspicious activity

**Procedure**:
1. Export full date range
2. Open in spreadsheet software (Excel, Google Sheets)
3. Filter by `username` column
4. Save filtered version as separate file

**Excel Filter Steps**:
```
1. Open CSV file in Excel
2. Select header row (Data → Filter)
3. Filter 'username' column for target user
4. Copy filtered results to new workbook
5. Save as "UserActivity_[USERNAME]_[DATES].csv"
```

---

## Post-Export Verification

### Step 1: Validate File Integrity

**Check File Size**:
```bash
# Verify file is not empty
ls -lh audit_logs_export.csv
# Expected: Non-zero file size (varies by date range)

# Count lines (excluding header)
line_count=$(tail -n +2 audit_logs_export.csv | wc -l)
echo "Exported $line_count audit log entries"
```

**Verify Format**:
```bash
# Check CSV structure
head -n 5 audit_logs_export.csv
# Should show: id,timestamp,user_id,username,action,ip_address,outcome,reason

# Verify no corrupt characters
file audit_logs_export.csv
# Expected: CSV text data or ASCII text
```

**Checksum Verification**:
```bash
# Generate SHA-256 checksum
shasum -a 256 audit_logs_export.csv > audit_logs_export.csv.sha256

# Verify later (for integrity confirmation)
shasum -c audit_logs_export.csv.sha256
# Expected: audit_logs_export.csv: OK
```

### Step 2: Validate Data Completeness

**Spot Check Critical Actions**:
```bash
# Verify system change audits are present
grep -c "SYSTEM_SETTINGS_UPDATED" audit_logs_export.csv
grep -c "BACKUP_CREATED" audit_logs_export.csv
grep -c "STOCK_ADJUSTMENT" audit_logs_export.csv

# Expected: Counts > 0 for active systems
```

**Date Range Validation**:
```bash
# Check first and last timestamps
head -n 2 audit_logs_export.csv | tail -n 1 | cut -d',' -f2
tail -n 1 audit_logs_export.csv | cut -d',' -f2

# Verify timestamps are within requested date range
```

### Step 3: Prepare Delivery Package

**Package Contents**:
```
audit_logs_package_YYYYMMDD/
├── AuditLogs_2026-01-01_to_2026-12-31.csv
├── AuditLogs_2026-01-01_to_2026-12-31.csv.sha256
├── README.txt (export description and metadata)
└── CERTIFICATE_OF_AUTHenticity.txt (optional, for legal purposes)
```

**README.txt Template**:
```
Audit Log Export Package
Generated: 2026-05-27
System: Simpo Pharmacy Management System v1.0
Export Period: 2026-01-01 to 2026-12-31

Contents:
- CSV file with complete audit log entries for specified period
- SHA-256 checksum for file integrity verification
- This README file

Export Details:
- Total Records: [COUNT]
- Date Range: [START_DATE] to [END_DATE]
- Actions Included: All system, user, and inventory actions
- Format: CSV (Comma-Separated Values)
- Encoding: UTF-8

Verification Instructions:
1. Verify file integrity: shasum -c [CSV_FILE].sha256
2. Open in spreadsheet software or text editor
3. Filter by 'action' column to find specific event types

Contact: support@simpopharmacy.com
```

---

## Badan POM Inspection Delivery

### Delivery Options

**Option 1: Digital Upload** (Preferred)
- Upload to secure Badan POM portal
- Use encrypted file transfer (SFTP, HTTPS)
- Provide checksum for integrity verification

**Option 2: Physical Media**
- Save to USB drive (encrypted)
- Provide written receipt of transfer
- Deliver in person to Badan POM office

**Option 3: Secure Email**
- Encrypt files (password-protected ZIP)
- Send password via separate channel
- Confirm delivery and receipt

### Documentation Package

**For Badan POM Inspectors**:
1. **Audit Log Export** - CSV/JSON files
2. **System Certificate** - Software version and configuration
3. **Retention Policy** - Proof of 5-year retention compliance
4. **Access Control Matrix** - RBAC permissions and user roles
5. **Data Integrity Report** - Checksums and verification results

**Cover Letter Template**:
```
[Pharmacy Letterhead]
Date: [Date]
To: Badan POM Inspection Team
Subject: Audit Log Export - Compliance Inspection

Dear Inspector,

Please find attached the audit log export for [Pharmacy Name]
covering the period [Start Date] to [End Date].

Package Contents:
- Audit log export: [Filename]
- SHA-256 checksum: [Checksum]
- Total records: [Count]
- Date range: [Date Range]

System Information:
- Software: Simpo Pharmacy Management System v1.0
- Database: PostgreSQL 14+
- Audit retention: 5+ years
- Append-only: Yes (enforced at database level)

The attached audit logs comply with Badan POM requirements:
✓ Append-only audit trail
✓ User identification (ID, username)
✓ Accurate timestamps (Asia/Jakarta timezone)
✓ Reason for all system changes
✓ 5-year minimum retention

Please contact [Name] at [Email] for any clarifications.

Sincerely,
[Name]
[Title]
[Pharmacy Name]
```

---

## Troubleshooting

### Export Fails or Times Out

**Symptom**: Export fails after 30 seconds or returns error

**Solutions**:
1. Reduce date range to less than 1 year
2. Use action filters to reduce result set
3. Try during off-peak hours (early morning)
4. Contact support for server timeout adjustment

### Empty Export File

**Symptom**: Export succeeds but file is empty or has no data rows

**Solutions**:
1. Verify date range has data (check via web UI first)
2. Confirm user role has access to audit logs
3. Check that start_date is before end_date
4. Ensure JWT token is valid and not expired

### Corrupted CSV File

**Symptom**: File won't open in Excel or shows garbled characters

**Solutions**:
1. Verify file encoding is UTF-8
2. Try alternative text editor (VS Code, Notepad++)
3. Re-export with JSON format instead
4. Check for special characters in reason field

### Missing Audit Entries

**Symptom**: Expected entries are missing from export

**Solutions**:
1. Verify date range covers expected period
2. Check for multiple exports (combine if needed)
3. Ensure system was operational during period
4. Verify database replication status

---

## Support & Resources

**Technical Support**:  
- Email: support@simpopharmacy.com  
- Phone: +62-21-1234-5678  
- Hours: Monday-Friday, 08:00-17:00 WIB

**Compliance Support**:  
- Email: compliance@simpopharmacy.com  
- Documentation: https://docs.simpopharmacy.com/compliance

**API Documentation**:  
- Swagger UI: https://api.simpopharmacy.com/swagger/index.html  
- OpenAPI Spec: https://api.simpopharmacy.com/swagger.yaml

---

## Revision History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-05-27 | Initial export procedure documentation | Development Team |

---

**Document Classification**: Internal - Operational  
**Distribution**: System Administrators, Pharmacy Owners, Compliance Officers
