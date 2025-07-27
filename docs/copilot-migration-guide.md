# GitHub Copilot Language Server Migration Guide

## Overview

This guide provides step-by-step instructions for migrating from `gopls` to the official GitHub Copilot language server in SuperOpenCode. The migration process is designed to be safe, reversible, and minimally disruptive to your development workflow.

## Migration Strategies

### Strategy 1: Gradual Migration (Recommended)

Best for: Most users, production environments, teams wanting to test Copilot gradually

**Benefits**:
- ✅ Zero downtime
- ✅ Easy rollback
- ✅ Test Copilot features while keeping gopls
- ✅ Minimal risk

**Process**: Enable Copilot alongside gopls, test features, then optionally replace gopls

### Strategy 2: Complete Replacement

Best for: Users committed to Copilot, new projects, development environments

**Benefits**:
- ✅ Simplified configuration
- ✅ Full Copilot feature access
- ✅ Reduced resource usage

**Considerations**:
- ⚠️ Potential feature gaps compared to gopls
- ⚠️ Requires comprehensive testing

### Strategy 3: Hybrid Mode

Best for: Power users, complex projects, teams using both traditional and AI-assisted development

**Benefits**:
- ✅ Best of both worlds
- ✅ Maximum flexibility
- ✅ Feature redundancy

**Considerations**:
- ⚠️ Higher resource usage
- ⚠️ Potential conflicts between servers

## Pre-Migration Checklist

### ✅ Prerequisites Verification

1. **Check GitHub Copilot Access**:
   ```bash
   # Visit GitHub Copilot settings
   open https://github.com/settings/copilot
   
   # Verify subscription status
   gh api user/copilot_billing
   ```

2. **Verify System Requirements**:
   ```bash
   # Check Node.js version (>= 20.8.0)
   node --version
   
   # Check npm version
   npm --version
   
   # Check available disk space (>100MB)
   df -h
   ```

3. **Backup Current Configuration**:
   ```bash
   # Backup configuration
   cp ~/.config/opencode/config.json ~/.config/opencode/config.json.backup
   
   # Or for project-specific config
   cp ./opencode.json ./opencode.json.backup
   ```

4. **Test Current Setup**:
   ```bash
   # Verify gopls is working
   opencode --test-lsp
   
   # Check current LSP server status
   opencode --lsp-status
   ```

### ✅ Environment Preparation

1. **Install GitHub CLI (Optional)**:
   ```bash
   # macOS
   brew install gh
   
   # Linux
   sudo apt install gh
   
   # Windows
   winget install --id GitHub.cli
   ```

2. **Authenticate with GitHub**:
   ```bash
   # Using GitHub CLI
   gh auth login
   
   # Or prepare personal access token
   # https://github.com/settings/tokens
   ```

## Migration Process

### Phase 1: Install and Configure Copilot

#### Step 1: Install Copilot Language Server

**Automatic Installation (Recommended)**:
```json
{
  "copilot": {
    "enable_copilot": true,
    "auto_install": true
  }
}
```

**Manual Installation**:
```bash
# Global installation
npm install -g @github/copilot-language-server

# Verify installation
copilot-language-server --version
```

#### Step 2: Basic Configuration

Create minimal configuration file:

```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": false,
    "fallback_to_gopls": true,
    "auth_token": ""
  }
}
```

#### Step 3: Authentication Setup

**Interactive Authentication**:
```bash
# Start SuperOpenCode - it will prompt for authentication
opencode

# Follow browser prompts to authenticate
```

**Token-based Authentication**:
```bash
# Set environment variable
export GITHUB_TOKEN=ghp_your_token_here

# Or add to configuration
```

```json
{
  "copilot": {
    "auth_token": "ghp_your_token_here"
  }
}
```

#### Step 4: Verify Installation

```bash
# Check Copilot status
opencode --check-copilot

# Verify language server is running
opencode --lsp-status
```

### Phase 2: Test Copilot Features

#### Step 5: Test Basic Functionality

1. **Open a Go file**:
   ```bash
   opencode main.go
   ```

2. **Test Completions**:
   - Start typing a function
   - Verify AI suggestions appear
   - Compare with gopls suggestions

3. **Test Chat Features**:
   ```bash
   # In SuperOpenCode
   :copilot chat "Explain this function"
   ```

#### Step 6: Feature Comparison Testing

Create a test checklist:

| Feature | gopls | Copilot | Notes |
|---------|-------|---------|-------|
| Code Completion | ✅ | ✅ | Test both |
| Go to Definition | ✅ | ✅ | Verify accuracy |
| Find References | ✅ | ✅ | Check completeness |
| Hover Documentation | ✅ | ✅ | Compare quality |
| Error Diagnostics | ✅ | ⚠️ | May differ |
| Refactoring | ✅ | ⚠️ | Limited in Copilot |
| AI Suggestions | ❌ | ✅ | New feature |
| Chat Assistance | ❌ | ✅ | New feature |

#### Step 7: Performance Testing

```bash
# Test completion speed
time opencode --test-completion

# Monitor resource usage
top -p $(pgrep copilot-language-server)

# Compare memory usage
ps aux | grep -E "(gopls|copilot)"
```

### Phase 3: Gradual Feature Migration

#### Step 8: Enable More Copilot Features

Gradually enable additional features:

```json
{
  "copilot": {
    "enable_copilot": true,
    "completion_enabled": true,
    "chat_enabled": true,
    "agent_config": {
      "coding_agent": true,
      "debugging_agent": true,
      "documentation_agent": false
    }
  }
}
```

#### Step 9: Adjust Performance Settings

Fine-tune for your workflow:

```json
{
  "copilot": {
    "performance": {
      "max_completion_time": 3000,
      "debounce_delay": 200,
      "max_parallel_requests": 5
    },
    "cache_enabled": true,
    "cache_size": 100
  }
}
```

### Phase 4: Complete Migration (Optional)

#### Step 10: Replace gopls (If Desired)

After thorough testing, optionally replace gopls:

```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": true
  },
  "lsp": {
    "disable_gopls": true
  }
}
```

#### Step 11: Clean Up Configuration

Remove unnecessary fallback options:

```json
{
  "copilot": {
    "enable_copilot": true,
    "replace_gopls": true,
    "fallback_to_gopls": false
  }
}
```

## Migration Validation

### Functional Validation

1. **Code Completion Test**:
   ```go
   package main
   
   func main() {
       // Type "fmt.Pr" and verify suggestions
   }
   ```

2. **Error Detection Test**:
   ```go
   package main
   
   func main() {
       undefinedVariable // Should show error
   }
   ```

3. **Go to Definition Test**:
   - Click on a function call
   - Verify it navigates to definition

4. **Chat Feature Test**:
   ```bash
   # In SuperOpenCode
   :copilot chat "How do I handle errors in Go?"
   ```

### Performance Validation

```bash
# Measure startup time
time opencode --version

# Test completion latency
opencode --benchmark-completions

# Check memory usage
opencode --memory-usage
```

### Integration Validation

```bash
# Test with existing projects
cd /path/to/go/project
opencode .

# Verify LSP features work
opencode --test-integration
```

## Rollback Procedures

### Emergency Rollback

If Copilot causes issues, quickly disable it:

```json
{
  "copilot": {
    "enable_copilot": false
  }
}
```

Or via environment variable:
```bash
export OPENCODE_COPILOT_ENABLE_COPILOT=false
opencode
```

### Complete Rollback

1. **Restore Original Configuration**:
   ```bash
   cp ~/.config/opencode/config.json.backup ~/.config/opencode/config.json
   ```

2. **Remove Copilot Language Server**:
   ```bash
   npm uninstall -g @github/copilot-language-server
   ```

3. **Clear Cached Data**:
   ```bash
   rm -rf ~/.cache/opencode/copilot
   ```

4. **Restart SuperOpenCode**:
   ```bash
   opencode
   ```

### Partial Rollback

Keep Copilot installed but disable specific features:

```json
{
  "copilot": {
    "enable_copilot": true,
    "completion_enabled": false,
    "chat_enabled": false,
    "replace_gopls": false,
    "fallback_to_gopls": true
  }
}
```

## Common Migration Scenarios

### Scenario 1: Team Migration

For teams, use a phased approach:

1. **Pilot Phase** (1-2 developers):
   ```json
   {
     "copilot": {
       "enable_copilot": true,
       "replace_gopls": false
     }
   }
   ```

2. **Gradual Rollout** (25% of team):
   - Share pilot feedback
   - Provide training
   - Monitor performance

3. **Full Deployment** (100% of team):
   - Standardize configuration
   - Update documentation
   - Provide support procedures

### Scenario 2: CI/CD Environment

For automated environments:

```json
{
  "copilot": {
    "enable_copilot": false
  }
}
```

Use environment variables for flexibility:
```bash
export OPENCODE_COPILOT_ENABLE_COPILOT=${ENABLE_COPILOT:-false}
```

### Scenario 3: Multiple Projects

Use project-specific configurations:

```bash
# Project A (using Copilot)
cd /project-a
echo '{"copilot": {"enable_copilot": true}}' > opencode.json

# Project B (using gopls)
cd /project-b
echo '{"copilot": {"enable_copilot": false}}' > opencode.json
```

## Troubleshooting Migration Issues

### Issue: Authentication Failures

**Symptoms**:
- "Authentication failed" errors
- Copilot features not working

**Solutions**:
1. Verify GitHub Copilot subscription
2. Check token permissions
3. Re-authenticate:
   ```bash
   opencode --copilot-reauth
   ```

### Issue: Performance Degradation

**Symptoms**:
- Slow completions
- High memory usage
- Unresponsive interface

**Solutions**:
1. Adjust performance settings:
   ```json
   {
     "copilot": {
       "performance": {
         "max_completion_time": 1000,
         "max_parallel_requests": 3
       }
     }
   }
   ```

2. Enable caching:
   ```json
   {
     "copilot": {
       "cache_enabled": true,
       "cache_size": 200
     }
   }
   ```

### Issue: Feature Conflicts

**Symptoms**:
- Duplicate completions
- Conflicting diagnostics

**Solutions**:
1. Use hybrid mode carefully:
   ```json
   {
     "lsp": {
       "enable_hybrid_mode": false
     }
   }
   ```

2. Choose primary server:
   ```json
   {
     "copilot": {
       "replace_gopls": true
     }
   }
   ```

### Issue: Installation Problems

**Symptoms**:
- Server not found
- Installation failures

**Solutions**:
1. Manual installation:
   ```bash
   npm install -g @github/copilot-language-server
   ```

2. Check permissions:
   ```bash
   npm config set prefix ~/.npm-global
   ```

3. Specify explicit path:
   ```json
   {
     "copilot": {
       "server_path": "/explicit/path/to/server"
     }
   }
   ```

## Post-Migration Best Practices

### 1. Regular Updates

```bash
# Update Copilot language server
npm update -g @github/copilot-language-server

# Check for configuration updates
opencode --check-config-updates
```

### 2. Monitoring and Metrics

```bash
# Monitor performance
opencode --performance-report

# Check usage statistics
opencode --copilot-stats
```

### 3. Team Training

- Provide Copilot usage guidelines
- Share best practices for AI-assisted development
- Create project-specific configuration templates

### 4. Security Considerations

```json
{
  "copilot": {
    "security": {
      "disable_telemetry": true,
      "private_mode": false
    }
  }
}
```

## Migration Timeline

### Week 1: Preparation
- ✅ Verify prerequisites
- ✅ Backup configurations
- ✅ Install Copilot language server
- ✅ Set up authentication

### Week 2: Testing
- ✅ Enable Copilot alongside gopls
- ✅ Test core features
- ✅ Performance benchmarking
- ✅ Team feedback collection

### Week 3: Gradual Adoption
- ✅ Enable additional features
- ✅ Fine-tune configuration
- ✅ Monitor performance
- ✅ Address issues

### Week 4: Full Migration (Optional)
- ✅ Replace gopls if desired
- ✅ Clean up configuration
- ✅ Update documentation
- ✅ Finalize rollback procedures

## Support and Resources

### Documentation
- **Configuration Reference**: `copilot-configuration-reference.md`
- **Installation Guide**: `copilot-installation-guide.md`
- **API Reference**: `copilot-api-reference.md`
- **Troubleshooting**: `copilot-troubleshooting.md`

### Community Support
- **GitHub Issues**: https://github.com/kirmad/superopencode/issues
- **GitHub Copilot Support**: https://support.github.com/
- **Discussion Forums**: Check project README for community links

### Professional Support
For enterprise deployments, consider:
- GitHub Enterprise support
- Professional services consultation
- Custom integration development

## Conclusion

The migration from gopls to GitHub Copilot language server enhances your development workflow with AI-powered assistance while maintaining the reliability of traditional language server features. Follow this guide systematically, test thoroughly, and don't hesitate to rollback if issues arise.

Remember: Migration is a process, not an event. Take your time to ensure a smooth transition that benefits your development workflow.