# Tasks: Export Telegram Group Members

**Feature**: Export Telegram Group Members
**Branch**: `001-export-telegram-members`
**Generated**: 2025-12-06

## Overview

This document breaks down the implementation of the Telegram group members export application into discrete, actionable tasks organized by user story priority. The application is a single Go file that authenticates with Telegram, retrieves group member information using group IDs, and exports it to JSON format.

## Implementation Strategy

- **MVP Scope**: User Story 1 (Export Telegram Group Members) - basic functionality to authenticate with Telegram, retrieve members from a specified group ID, and export to JSON
- **Delivery Approach**: Implement foundational components first, then progressively add user stories in priority order (P1, P2, P3)
- **Parallel Opportunities**: Configuration, data structures, and client initialization can proceed in parallel with core functionality

---

## Phase 1: Project Setup

**Goal**: Establish foundational project structure and dependencies for Telegram group member export application

- [X] T001 Create go.mod file with module name telegram-exporter
- [X] T002 Install and add dependencies github.com/celestix/gotgproto v1.0.0-rc.3, github.com/xelaj/mtproto, os.Getenv for config
- [X] T003 Create empty main.go file with package declaration and main function
- [X] T004 Create .env.example file with PHONE, API_ID, API_HASH, GROUP_ID, PROXY_HOST, PROXY_PORT, PROXY_SECRET
- [X] T005 Create out/ directory in project root for output files
- [X] T006 Update .gitignore to exclude .env, *.session, and out/ files

## Phase 2: Foundational Components

**Goal**: Implement core configuration, data structures, and authentication required for all user stories

- [X] T007 [P] Create GroupMember struct in main.go with ID (int64), Username (string), FirstName (string), LastName (string)
- [X] T008 [P] Create ExportConfiguration struct in main.go with APIID, APIHash, PhoneNumber, GroupID, OutputDir, ProxyHost, ProxyPort, ProxySecret
- [X] T009 [P] Create LoadConfig function in main.go to read env vars into ExportConfiguration
- [X] T010 [P] Create ValidateConfig function in main.go to validate required configuration fields
- [X] T011 [P] Create ExportResult struct in main.go with GroupID, GroupName, MemberCount, OutputFilePath, ExportTime, Success, ErrorMessage
- [X] T012 [P] Initialize Telegram client using gotgproto with configuration values
- [X] T013 [P] Implement authentication flow with code request and sign-in handling

## Phase 3: [US1] Export Telegram Group Members

**Goal**: Implement core functionality to export Telegram group members to JSON file as specified in User Story 1

**Independent Test**: The application can be tested by connecting to a Telegram group and successfully exporting its members to a JSON file that contains all required member information fields.

- [X] T014 [US1] Create GetGroupMembers function to retrieve all members from specified group ID using Telegram client
- [X] T015 [US1] Implement pagination handling for groups with large numbers of members
- [X] T016 [US1] Create ExportToJSON function to write GroupMember slice to JSON file with timestamp format
- [X] T017 [US1] Implement file path generation using group name and current date in format 'groupname_members_YYYYMMDD.json'
- [X] T018 [US1] Create directory for output file if it doesn't exist
- [X] T019 [US1] Implement main export workflow connecting configuration → authentication → group members → JSON export
- [X] T020 [US1] Add progress feedback during member retrieval with count and percentage
- [X] T021 [US1] Test with actual group to verify 100% of accessible members are included
- [X] T022 [US1] Verify JSON output follows standard formatting for consumption by other tools

## Phase 4: [US2] Select Specific Group for Export

**Goal**: Enable users to select which specific group to export members from, supporting multiple Telegram groups

**Independent Test**: The application presents a list of accessible groups for the user to select from, and exports members from the chosen group.

- [X] T023 [US2] Create GetAccessibleGroups function to retrieve all groups accessible to the authenticated user
- [X] T024 [US2] Implement group selection interface (interactive prompt or CLI argument) to choose target group
- [X] T025 [US2] Modify configuration to accept either group ID or group name for selection
- [X] T026 [US2] Add validation to ensure selected group exists and user has access
- [X] T027 [US2] Update authentication flow to use selected group for export

## Phase 5: [US3] Export Custom Member Information Fields

**Goal**: Allow users to select which specific member information fields to export for privacy and efficiency

**Independent Test**: The application allows users to select which member attributes to include in the export (e.g., username, user ID, first name, last name).

- [X] T028 [US3] Extend GroupMember struct to include additional fields: IsBot, IsScam, IsFake, PhoneNumber
- [X] T029 [US3] Create ExportOptions struct to specify which fields to include in export
- [X] T030 [US3] Modify ExportToJSON function to support configurable field selection
- [X] T031 [US3] Implement configuration of export options via environment variables or CLI flags
- [X] T032 [US3] Update JSON serialization to include only selected fields

## Phase 6: Error Handling & Edge Cases

**Goal**: Implement robust error handling and address edge cases identified in the specification

- [X] T033 Create error handling for authentication failures with user-friendly messages
- [X] T034 Implement error handling for invalid group IDs or inaccessible groups
- [X] T035 Add error handling for rate limiting by Telegram API with appropriate delays
- [X] T036 Handle privacy-restricted member information gracefully (empty fields instead of errors)
- [X] T037 Implement timeout handling for network requests
- [X] T038 Add retry logic for transient network failures
- [X] T039 Create logging mechanism for debugging and user feedback

## Phase 7: Polish & Cross-Cutting Concerns

**Goal**: Finalize the application with polish, documentation, and optimization

- [X] T040 Optimize memory usage for large groups (streaming instead of loading all members in memory)
- [X] T041 Add performance timing to ensure export completes within 5 minutes for 10,000 members
- [X] T042 Create comprehensive README with setup and usage instructions
- [ ] T043 Implement graceful shutdown handling
- [ ] T044 Add usage examples to main.go comments
- [ ] T045 Final testing of all implemented user stories
- [ ] T046 Verify compliance with Telegram's API terms of service

---

## Dependencies

**User Story Completion Order**:
1. User Story 1 (Export Members) - Foundation for all other stories
2. User Story 2 (Select Group) - Depends on User Story 1
3. User Story 3 (Custom Fields) - Can build on User Story 1

**Task Dependencies**:
- T001-006 must complete before any other tasks
- T007-013 must complete before T014 and other US1 tasks
- T014 requires completed configuration and authentication (T007-T013)
- T023-025 require authentication flow to be working (US1 completed)

## Parallel Execution Examples

**Per User Story**:
- User Story 1: T014 (GetGroupMembers) and T016 (ExportToJSON) can be developed in parallel after T007-T013 are complete
- User Story 2: T023 (GetAccessibleGroups) and T024 (Group selection interface) can be developed in parallel
- User Story 3: T028 (Extended GroupMember) and T029 (ExportOptions) can be developed in parallel

**Cross-Story**:
- Configuration and data structures (Phase 2) can be developed in parallel with setup tasks but must complete before user story work
- Error handling (Phase 6) can be added incrementally throughout development
- Polish tasks (Phase 7) can be worked on after core functionality (US1) is complete