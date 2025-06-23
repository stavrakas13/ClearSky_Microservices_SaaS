// front-end/app.js
import express           from 'express';
import session           from 'express-session';
import cookieParser      from 'cookie-parser'; // Add cookie parser import
import path              from 'path';
import { fileURLToPath } from 'url';
import morgan            from 'morgan';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app       = express();

// ─────────────────────────────────────────────────────────────────────────────
// 0)  API base-URL resolution
// ─────────────────────────────────────────────────────────────────────────────
const API_BASE        =
  process.env.GO_API_URL       ||
  process.env.ORCHESTRATOR_URL ||
  'http://orchestrator:8080';

const GOOGLE_AUTH_URL =
  process.env.GOOGLE_AUTH_URL ||
  'http://google_auth_service:8086';    // Use Docker service name

// ─────────────────────────────────────────────────────────────────────────────
// 1)  3rd-party middleware
// ─────────────────────────────────────────────────────────────────────────────
app.use(morgan('dev'));
app.use(cookieParser()); // Add cookie parser middleware

// ─────────────────────────────────────────────────────────────────────────────
// 2)  Static assets
// ─────────────────────────────────────────────────────────────────────────────
app.use(express.static(path.join(__dirname, 'public')));

// ─────────────────────────────────────────────────────────────────────────────
// 3)  Body-parsers
// ─────────────────────────────────────────────────────────────────────────────
app.use(express.urlencoded({ extended: false }));
app.use(express.json());

// ─────────────────────────────────────────────────────────────────────────────
// 4)  Sessions & locals
// ─────────────────────────────────────────────────────────────────────────────
app.use(session({
  secret           : 'change-this-secret',
  resave           : false,
  saveUninitialized: true,
}));
app.use((req, res, next) => {
  res.locals.user       = req.session.user || null;
  res.locals.currentUrl = req.originalUrl;
  res.locals.API_BASE   = API_BASE;
  next();
});

// ─────────────────────────────────────────────────────────────────────────────
// 5)  EJS templating
// ─────────────────────────────────────────────────────────────────────────────
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

// ─────────────────────────────────────────────────────────────────────────────
// 6)  GOOGLE OAUTH PROXY
//    Forward front-end `/auth/google/...` to your google_auth_service.
// ─────────────────────────────────────────────────────────────────────────────
app.get('/auth/google/login', (req, res) => {
  const role = req.query.role || 'institution_representative';
  // For Docker, we need to redirect to the external URL
  const externalGoogleAuthUrl = process.env.GOOGLE_AUTH_EXTERNAL_URL || 'http://localhost:8086';
  res.redirect(`${externalGoogleAuthUrl}/auth/google/login?role=${role}`);
});

// Handle successful Google login callback
app.get('/auth/google/callback', (req, res) => {
  console.log('Google callback received:', req.query);
  
  // Check if we have a JWT cookie
  const token = req.cookies.token;
  const role = req.query.role || 'institution_representative';
  const email = req.query.email;
  
  if (token && req.query.google_login === 'success') {
    // Set session for Google user
    req.session.user = {
      username: email || 'google_user',
      role: role
    };
    
    console.log('Set session user:', req.session.user);
    
    // Redirect based on role
    let redirectPath = '/';
    switch (role) {
      case 'student':
        redirectPath = '/student';
        break;
      case 'instructor':
        redirectPath = '/instructor';
        break;
      case 'institution_representative':
        redirectPath = '/institution';
        break;
      default:
        redirectPath = '/';
    }
    
    res.redirect(redirectPath);
  } else {
    console.log('Google login failed - no token or success flag');
    res.redirect('/login?error=google_login_failed');
  }
});

// ─────────────────────────────────────────────────────────────────────────────
// 7)  Auth helper
// ─────────────────────────────────────────────────────────────────────────────
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect('/login');
    if (role) {
      if (role === 'institution') {
        if (!['institution','representative','institution_representative']
              .includes(req.session.user.role)) {
          return res.redirect(`/${req.session.user.role}`);
        }
      } else if (req.session.user.role !== role) {
        return res.redirect(`/${req.session.user.role}`);
      }
    }
    next();
  };
}

// ─────────────────────────────────────────────────────────────────────────────
// 8)  Dummy users (dev only)
// ─────────────────────────────────────────────────────────────────────────────
const users = { alice: 'student', bob: 'instructor', iris: 'institution' };

// ─────────────────────────────────────────────────────────────────────────────
// 9)  UI routes
// ─────────────────────────────────────────────────────────────────────────────

// Home
app.get('/', (req, res) => {
  if (!req.session.user) return res.redirect('/login');
  res.redirect(`/${req.session.user.role}`);
});

// Signup / Login
app.get('/signup', (_, res) =>
  res.render('signup', { title: 'Sign Up', user: null })
);

app.get('/login', (_, res) =>
  res.render('login', { title: 'Log in', error: null, user: null })
);

// CLASSIC form POST – creates session
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  try {
    const response = await fetch(`${API_BASE}/user/login`, {
      method : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body   : JSON.stringify({ username, password })
    });
    const data = await response.json();

    if (!response.ok || !data.role) {
      return res.render('login', {
        title : 'Log in',
        error : data.message || 'Invalid credentials',
        user  : null
      });
    }

    req.session.user = { username, role: data.role };

    if (['institution_representative','representative']
        .includes(data.role))      return res.redirect('/institution');
    else if (data.role === 'instructor') return res.redirect('/instructor');
    else if (data.role === 'student')    return res.redirect('/student');
    else                                 return res.redirect('/');
  } catch (err) {
    return res.render('login', {
      title : 'Log in',
      error : 'Login failed',
      user  : null
    });
  }
});

app.post('/api/session', (req, res) => {
  const { username, role } = req.body;
  if (!username || !role) {
    return res.status(400).json({ error: 'username and role required' });
  }
  req.session.user = { username, role };
  res.sendStatus(200);
});

app.get('/logout', (req, res) =>
  req.session.destroy(() => res.redirect('/login'))
);

// Student UI
app.get('/student',            auth('student'), (req,res)=>res.render('student/dashboard',    { user:req.session.user, title:'Dashboard' }));
app.get('/student/statistics', auth('student'), (req,res)=>res.render('student/statistics',   { user:req.session.user, title:'Statistics' }));
app.get('/student/my-courses', auth('student'), (req,res)=>res.render('student/myCourses',    { user:req.session.user, title:'My Courses' }));
app.get('/student/request',    auth('student'), (req,res)=>res.render('student/reviewRequest',{ user:req.session.user, title:'Review Request' }));
app.get('/student/status',     auth('student'), (req,res)=>res.render('student/reviewStatus', { user:req.session.user, title:'Review Status' }));
app.get('/student/personal',   auth('student'), (req,res)=>res.render('student/personal',     { user:req.session.user, title:'Personal Grades' }));

// Instructor UI
app.get('/instructor',              auth('instructor'), (req,res)=>res.render('instructor/dashboard', { user:req.session.user, title:'Dashboard' }));
app.get('/instructor/post-initial', auth('instructor'), (req,res)=>res.render('instructor/postInitial',{ user:req.session.user, title:'Post Initial' }));
app.get('/instructor/post-final',   auth('instructor'), (req,res)=>res.render('instructor/postFinal',  { user:req.session.user, title:'Post Final' }));
app.get('/instructor/review-list',  auth('instructor'), (req,res)=>res.render('instructor/reviewList', { user:req.session.user, title:'Review Requests' }));
app.get('/instructor/reply',        auth('instructor'), (req,res)=>{
  // Extract query parameters from URL
  const course_id   = req.query.course   || '';
  const exam_period = req.query.period   || '';
  const user_id     = req.query.student  || '';
  // Optionally, you could look up course_name/student_name from DB if needed

  res.render('instructor/replyForm',{
    user        : req.session.user,
    title       : 'Reply to Review Request',
    request_id  : '', // not used, but kept for compatibility
    course_name : course_id,
    exam_period : exam_period,
    student_name: user_id,
  });
});
app.get('/instructor/statistics',   auth('instructor'), (req,res)=>res.render('instructor/statistics', { user:req.session.user, title:'Statistics' }));

// Institution UI
app.get('/institution',                 auth('institution'), (req,res)=>res.render('institution/dashboard',      { user:req.session.user, title:'Dashboard' }));
app.get('/institution/register',        auth('institution'), (req,res)=>res.render('institution/register',       { user:req.session.user, title:'Register' }));
app.get('/institution/purchase',        auth('institution'), (req,res)=>res.render('institution/purchase',       { user:req.session.user, title:'Purchase' }));
app.get('/institution/user-management', auth('institution'), (req,res)=>res.render('institution/userManagement', { user:req.session.user, title:'Users' }));
app.get('/institution/statistics',      auth('institution'), (req,res)=>res.render('institution/statistics',     { user:req.session.user, title:'Statistics' }));

// ─────────────────────────────────────────────────────────────────────────────
// 10) Server start-up
// ─────────────────────────────────────────────────────────────────────────────
const PORT = process.env.PORT || 3000;
app.listen(PORT, () =>
  console.log(`✔ Front-end listening at http://localhost:${PORT}`)
);

